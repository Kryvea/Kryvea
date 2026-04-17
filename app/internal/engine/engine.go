package engine

import (
	"github.com/Kryvea/Kryvea/internal/api"
	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog"
)

type Engine struct {
	addr        string
	rootPath    string
	mongo       *mongo.Driver
	levelWriter *zerolog.LevelWriter
}

func NewEngine(addr, rootPath, mongoURI, adminUser, adminPass string, levelWriter *zerolog.LevelWriter) (*Engine, error) {
	mongo, err := mongo.NewDriver(mongoURI, adminUser, adminPass, levelWriter)
	if err != nil {
		return nil, err
	}

	return &Engine{
		addr:        addr,
		rootPath:    rootPath,
		mongo:       mongo,
		levelWriter: levelWriter,
	}, nil
}

func (e *Engine) Serve() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		// TODO: this is a temporary solution to allow large files
		BodyLimit: 10000 * 1024 * 1024,

		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	logger := zerolog.New(*e.levelWriter).With().
		Str("source", "fiber-engine").
		Timestamp().Logger()

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))

	api := api.NewDriver(e.mongo, e.levelWriter)

	apiGroup := app.Group(util.JoinUrlPath(e.rootPath, "api"))
	apiGroup.Use(api.SessionMiddleware)
	{
		apiGroup.Get("/customers", api.GetCustomers)
		apiGroup.Get("/customers/:customer", api.GetCustomer)
		apiGroup.Post("/customers/:customer/templates/upload", api.AddCustomerTemplate)

		apiGroup.Get("/assessments", api.SearchAssessments)
		apiGroup.Get("/customers/:customer/assessments", api.GetAssessmentsByCustomer)
		apiGroup.Get("/assessments/owned", api.GetOwnedAssessments)
		apiGroup.Get("/assessments/:assessment", api.GetAssessment)
		apiGroup.Post("/assessments", api.AddAssessment)
		apiGroup.Patch("/assessments/:assessment", api.UpdateAssessment)
		apiGroup.Patch("/assessments/:assessment/status", api.UpdateAssessmentStatus)
		apiGroup.Delete("/assessments/:assessment", api.DeleteAssessment)
		apiGroup.Post("/assessments/:assessment/clone", api.CloneAssessment)
		apiGroup.Post("/assessments/:assessment/export", api.ExportAssessment)

		apiGroup.Get("/customers/:customer/targets", api.GetTargetsByCustomer)
		apiGroup.Get("/targets/:target", api.GetTarget)
		apiGroup.Post("/targets", api.AddTarget)
		apiGroup.Patch("/targets/:target", api.UpdateTarget)
		apiGroup.Delete("/targets/:target", api.DeleteTarget)

		apiGroup.Get("/categories/search", api.SearchCategories)
		apiGroup.Get("/categories", api.GetCategories)
		apiGroup.Get("/categories/:category", api.GetCategory)

		apiGroup.Get("/templates", api.GetTemplates)
		apiGroup.Get("/templates/:template", api.GetTemplate)
		apiGroup.Delete("/templates/:template", api.DeleteTemplate)

		apiGroup.Get("/vulnerabilities/user", api.GetUserVulnerabilities)
		apiGroup.Get("/vulnerabilities/search", api.SearchVulnerabilities)
		apiGroup.Get("/vulnerabilities/:vulnerability", api.GetVulnerability)
		apiGroup.Get("/assessments/:assessment/vulnerabilities", api.GetVulnerabilitiesByAssessment)
		apiGroup.Post("/vulnerabilities", api.AddVulnerability)
		apiGroup.Post("/vulnerabilities/:vulnerability/copy", api.CopyVulnerability)
		apiGroup.Put("/vulnerabilities/:vulnerability", api.UpdateVulnerability)
		apiGroup.Delete("/vulnerabilities/:vulnerability", api.DeleteVulnerability)
		apiGroup.Post("/assessments/:assessment/upload", api.ImportVulnerabilities)

		apiGroup.Get("/vulnerabilities/:vulnerability/pocs", api.GetPocsByVulnerability)
		apiGroup.Put("/vulnerabilities/:vulnerability/pocs", api.UpsertPocs)

		apiGroup.Get("/files/images/:file", api.GetImage)
		apiGroup.Get("/files/templates/:file", api.GetTemplateFile)
		apiGroup.Get("/files/customers/:file", api.GetCustomerImage)

		apiGroup.Get("/users/names", api.GetUsernames)
		apiGroup.Get("/users/me", api.GetMe)
		apiGroup.Patch("/users/me", api.UpdateMe)
		apiGroup.Patch("/users/me/assessments", api.UpdateOwnedAssessment)

		apiGroup.Post("/password/reset", api.ResetPassword)

		apiGroup.Post("/logout", api.Logout)

		// endpoints that don't require authentication
		apiGroup.Post("/login", api.Login)
	}

	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(api.AdminMiddleware)
	{
		adminGroup.Post("/customers", api.AddCustomer)
		adminGroup.Patch("/customers/:customer", api.UpdateCustomer)
		adminGroup.Put("/customers/:customer/logo", api.UpdateCustomerLogo)
		adminGroup.Delete("/customers/:customer", api.DeleteCustomer)

		adminGroup.Get("/categories/export", api.ExportCategories)
		adminGroup.Post("/categories", api.AddCategory)
		adminGroup.Post("/categories/upload", api.UploadCategories)
		adminGroup.Patch("/categories/:category", api.UpdateCategory)
		adminGroup.Delete("/categories/:category", api.DeleteCategory)

		adminGroup.Post("/templates/upload", api.AddGlobalTemplate)

		adminGroup.Get("/users", api.GetUsers)
		adminGroup.Get("/users/:user", api.GetUser)
		adminGroup.Post("/users", api.AddUser)
		adminGroup.Post("/users/:user/reset-password", api.ResetUserPassword)
		adminGroup.Patch("/users/:user", api.UpdateUser)
		adminGroup.Delete("/users/:user", api.DeleteUser)

		adminGroup.Get("/logs", api.GetLog)

		adminGroup.Get("/settings", api.GetSettings)
		adminGroup.Put("/settings", api.UpdateSettings)
	}

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	})

	logger.Info().Msg("Listening for connections on http://" + e.addr)
	if err := app.Listen(e.addr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
