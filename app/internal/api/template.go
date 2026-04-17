package api

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/bytedance/sonic"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type templateRequestData struct {
	Name       string `json:"name"`
	Language   string `json:"language"`
	Identifier string `json:"identifier"`
}

func (d *Driver) addTemplate(c *fiber.Ctx, ctx context.Context) (*mongo.Template, string) {
	// parse request data
	data := templateRequestData{}
	err := sonic.Unmarshal([]byte(c.FormValue("data")), &data)
	if err != nil {
		return nil, "Cannot parse JSON"
	}

	// validate request data
	errStr := d.validateTemplateData(&data)
	if errStr != "" {
		return nil, errStr
	}

	// parse template data from form
	templateData, filename, err := d.formDataReadFile(c, "template")
	if err != nil {
		return nil, "Cannot read template data"
	}

	if len(templateData) == 0 {
		return nil, "Template data is empty"
	}

	// check if the template mimetype is supported
	mimeType := mimetype.Detect(templateData)
	templateType, exists := mongo.SupportedTemplateMimeTypes[mimeType.String()]
	if !exists {
		return nil, "Invalid template type"
	}

	// insert file into the database
	fileID, mime, err := d.mongo.FileReference().Insert(ctx, templateData)
	if err != nil {
		return nil, "Cannot upload template"
	}

	// create a new template
	template := &mongo.Template{
		Name:         data.Name,
		Filename:     filename,
		Language:     data.Language,
		TemplateType: templateType,
		MimeType:     mime,
		Identifier:   data.Identifier,
		FileID:       fileID,
		Customer: &mongo.Customer{
			Model: mongo.Model{
				ID: uuid.Nil,
			},
		},
	}
	return template, ""
}

func (d *Driver) AddGlobalTemplate(c *fiber.Ctx) error {
	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	templateID, err := session.WithTransaction(func(ctx context.Context) (any, error) {
		// upload template file into database
		template, errStr := d.addTemplate(c, ctx)
		if errStr != "" {
			return uuid.Nil, errors.New(errStr)
		}

		// insert the template into the database
		templateID, err := d.mongo.Template().Insert(ctx, template)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return uuid.Nil, errors.New("Template with provided data already exists")
			}
			return uuid.Nil, errors.New("Cannot create template")
		}

		return templateID, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Template created",
		"template_id": templateID.(uuid.UUID),
	})
}

func (d *Driver) AddCustomerTemplate(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// if customer is specified check if user has access to it
	customer, errStr := d.customerFromParam(c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	templateID, err := session.WithTransaction(func(ctx context.Context) (any, error) {
		// upload template file into database
		template, errStr := d.addTemplate(c, ctx)
		if errStr != "" {
			return uuid.Nil, errors.New(errStr)
		}

		template.Customer.ID = customer.ID

		// insert the template into the database
		templateID, err := d.mongo.Template().Insert(ctx, template)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return uuid.Nil, errors.New("Template with provided data already exists")
			}
			return uuid.Nil, errors.New("Cannot create template")
		}

		return templateID, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Template created",
		"template_id": templateID.(uuid.UUID),
	})
}

func (d *Driver) GetTemplate(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// get template from param
	template, errStr := d.templateFromParam(c.Params("template"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to the template
	if !mongo.IsNullCustomer(template.Customer) && !user.CanAccessCustomer(template.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(template)
}

func (d *Driver) GetTemplates(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// get all templates
	templates, err := d.mongo.Template().GetAll(context.Background())
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Failed to fetch templates",
		})
	}

	// filter templates by user access
	filteredTemplates := []mongo.Template{}
	for _, template := range templates {
		if mongo.IsNullCustomer(template.Customer) || user.CanAccessCustomer(template.Customer.ID) {
			filteredTemplates = append(filteredTemplates, template)
		}
	}

	c.Status(fiber.StatusOK)
	return c.JSON(filteredTemplates)
}

func (d *Driver) DeleteTemplate(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// get template from param
	template, errStr := d.templateFromParam(c.Params("template"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to the template
	if (mongo.IsNullCustomer(template.Customer) && user.Role != mongo.RoleAdmin) ||
		(!mongo.IsNullCustomer(template.Customer) && !user.CanAccessCustomer(template.Customer.ID)) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	_, err = session.WithTransaction(func(ctx context.Context) (any, error) {
		// delete the template from the database
		err := d.mongo.Template().Delete(ctx, template.ID)
		if err != nil {
			return nil, errors.New("Failed to delete template")
		}

		return nil, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Template deleted",
	})
}

func (d *Driver) validateTemplateData(data *templateRequestData) string {
	if data.Name == "" {
		return "Name is required"
	}

	if data.Language == "" {
		return "Language is required"
	}

	return ""
}

func (d *Driver) templateFromParam(param string) (*mongo.Template, string) {
	if param == "" {
		return nil, "Template ID is required"
	}

	templateID, err := util.ParseUUID(param)
	if err != nil {
		return nil, "Invalid template ID"
	}

	template, err := d.mongo.Template().GetByID(context.Background(), templateID)
	if err != nil {
		return nil, "Invalid template ID"
	}

	return template, ""
}
