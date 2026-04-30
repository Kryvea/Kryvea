package api

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
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

func (d *Driver) addTemplate(c *fiber.Ctx, ctx context.Context) (*model.Template, string) {
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
	templateType, exists := model.SupportedTemplateMimeTypes[mimeType.String()]
	if !exists {
		return nil, "Invalid template type"
	}

	// insert file into the database
	fileID, mime, err := d.db.FileReference().Insert(ctx, templateData)
	if err != nil {
		return nil, "Cannot upload template"
	}

	// create a new template
	template := &model.Template{
		Name:         data.Name,
		Filename:     filename,
		Language:     data.Language,
		TemplateType: templateType,
		MimeType:     mime,
		Identifier:   data.Identifier,
		FileID:       fileID,
		Customer: &model.Customer{
			Model: model.Model{
				ID: uuid.Nil,
			},
		},
	}
	return template, ""
}

func (d *Driver) AddGlobalTemplate(c *fiber.Ctx) error {
	templateID, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		// upload template file into database
		template, errStr := d.addTemplate(c, ctx)
		if errStr != "" {
			return uuid.Nil, errors.New(errStr)
		}

		// insert the template into the database
		templateID, err := d.db.Template().Insert(ctx, template)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateKey) {
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
	user := c.Locals("user").(*model.User)

	// if customer is specified check if user has access to it
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
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

	templateID, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		// upload template file into database
		template, errStr := d.addTemplate(c, ctx)
		if errStr != "" {
			return uuid.Nil, errors.New(errStr)
		}

		template.Customer.ID = customer.ID

		// insert the template into the database
		templateID, err := d.db.Template().Insert(ctx, template)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateKey) {
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
	user := c.Locals("user").(*model.User)

	// get template from param
	template, errStr := d.templateFromParam(c.UserContext(), c.Params("template"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to the template
	if !model.IsNullCustomer(template.Customer) && !user.CanAccessCustomer(template.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(template)
}

func (d *Driver) GetTemplates(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// get all templates
	templates, err := d.db.Template().GetAll(c.UserContext())
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Failed to fetch templates",
		})
	}

	// filter templates by user access
	filteredTemplates := []model.Template{}
	for _, template := range templates {
		if model.IsNullCustomer(template.Customer) || user.CanAccessCustomer(template.Customer.ID) {
			filteredTemplates = append(filteredTemplates, template)
		}
	}

	c.Status(fiber.StatusOK)
	return c.JSON(filteredTemplates)
}

func (d *Driver) DeleteTemplate(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// get template from param
	template, errStr := d.templateFromParam(c.UserContext(), c.Params("template"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to the template
	if (model.IsNullCustomer(template.Customer) && user.Role != model.RoleAdmin) ||
		(!model.IsNullCustomer(template.Customer) && !user.CanAccessCustomer(template.Customer.ID)) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	if err := d.db.Template().Delete(c.UserContext(), template.ID); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Failed to delete template",
		})
	}

	d.gcFilesAsync()
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

func (d *Driver) templateFromParam(ctx context.Context, param string) (*model.Template, string) {
	if param == "" {
		return nil, "Template ID is required"
	}

	templateID, err := util.ParseUUID(param)
	if err != nil {
		return nil, "Invalid template ID"
	}

	template, err := d.db.Template().GetByID(ctx, templateID)
	if err != nil {
		return nil, "Invalid template ID"
	}

	return template, ""
}
