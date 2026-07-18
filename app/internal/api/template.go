package api

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
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
	data := templateRequestData{}
	err := sonic.Unmarshal([]byte(c.FormValue("data")), &data)
	if err != nil {
		return nil, "Cannot parse JSON"
	}

	errStr := d.validateTemplateData(&data)
	if errStr != "" {
		return nil, errStr
	}

	templateData, filename, err := d.formDataReadFile(c, "template")
	if err != nil {
		return nil, "Cannot read template data"
	}

	if len(templateData) == 0 {
		return nil, "Template data is empty"
	}

	mimeType := mimetype.Detect(templateData)
	templateType, exists := model.SupportedTemplateMimeTypes[mimeType.String()]
	if !exists {
		return nil, "Invalid template type"
	}

	fileID, mime, err := d.db.FileReference().Insert(ctx, templateData)
	if err != nil {
		return nil, "Cannot upload template"
	}

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

// insertTemplate uploads the template file and inserts the template record,
// optionally bound to a customer (uuid.Nil means global).
func (d *Driver) insertTemplate(c *fiber.Ctx, customerID uuid.UUID) error {
	templateID, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		template, errStr := d.addTemplate(c, ctx)
		if errStr != "" {
			return uuid.Nil, errors.New(errStr)
		}

		template.Customer.ID = customerID

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
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Template created",
		"template_id": templateID.(uuid.UUID),
	})
}

func (d *Driver) AddGlobalTemplate(c *fiber.Ctx) error {
	return d.insertTemplate(c, uuid.Nil)
}

func (d *Driver) AddCustomerTemplate(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if !user.CanAccessCustomer(customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	return d.insertTemplate(c, customer.ID)
}

func (d *Driver) GetTemplate(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	template, errStr := d.templateFromParam(c.UserContext(), c.Params("template"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if !model.IsNullCustomer(template.Customer) && !user.CanAccessCustomer(template.Customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(template)
}

func (d *Driver) GetTemplates(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	templates, err := d.db.Template().GetAll(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Failed to fetch templates")
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

	template, errStr := d.templateFromParam(c.UserContext(), c.Params("template"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if (model.IsNullCustomer(template.Customer) && user.Role != model.RoleAdmin) ||
		(!model.IsNullCustomer(template.Customer) && !user.CanAccessCustomer(template.Customer.ID)) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	if err := d.db.Template().Delete(c.UserContext(), template.ID); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Failed to delete template")
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
	return fromParam(ctx, param, "Template", d.db.Template().GetByID)
}
