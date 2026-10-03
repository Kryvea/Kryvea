package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type categoryRequestData struct {
	Identifier         string            `json:"identifier"`
	Name               string            `json:"name"`
	Subcategory        string            `json:"subcategory"`
	GenericDescription map[string]string `json:"generic_description"`
	GenericRemediation map[string]string `json:"generic_remediation"`
	LanguagesOrder     []string          `json:"languages_order"`
	References         []string          `json:"references"`
	Source             string            `json:"source"`
}

func (d *Driver) AddCategory(c *fiber.Ctx) error {
	data := &categoryRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr := d.validateCategoryData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	category := categoryFromRequestData(data)

	categoryID, err := d.db.Category().Insert(c.UserContext(), category)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateKey) {
			return jsonError(c, fiber.StatusBadRequest, categoryExistsMessage(category))
		}

		return jsonError(c, fiber.StatusBadRequest, "Cannot create category")
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Category created",
		"category_id": categoryID,
	})
}

func (d *Driver) UpdateCategory(c *fiber.Ctx) error {
	category, errStr := d.categoryFromParam(c.UserContext(), c.Params("category"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	data := &categoryRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr = d.validateCategoryData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	newCategory := categoryFromRequestData(data)

	err := d.db.Category().Update(c.UserContext(), category.ID, newCategory)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateKey) {
			return jsonError(c, fiber.StatusBadRequest, categoryExistsMessage(newCategory))
		}

		return jsonError(c, fiber.StatusBadRequest, "Cannot update category")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Category updated",
	})
}

func (d *Driver) DeleteCategory(c *fiber.Ctx) error {
	category, errStr := d.categoryFromParam(c.UserContext(), c.Params("category"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if err := d.db.Category().Delete(c.UserContext(), category.ID); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot delete category")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Category deleted",
	})
}

func (d *Driver) SearchCategories(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return jsonError(c, fiber.StatusBadRequest, "Query is required")
	}

	categories, err := d.db.Category().Search(c.UserContext(), query)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot search categories")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(categories)
}

func (d *Driver) GetCategories(c *fiber.Ctx) error {
	categories, err := d.db.Category().GetAll(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get categories")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(categories)
}

func (d *Driver) ExportCategories(c *fiber.Ctx) error {
	categories, err := d.db.Category().GetAll(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get categories")
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Disposition", "attachment; filename=categories.json")
	return c.JSON(categories)
}

func (d *Driver) GetCategory(c *fiber.Ctx) error {
	category, errStr := d.categoryFromParam(c.UserContext(), c.Params("category"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	c.Status(fiber.StatusOK)
	return c.JSON(category)
}

func (d *Driver) UploadCategories(c *fiber.Ctx) error {
	override := c.FormValue("override")

	dataBytes, err := util.ParseFormFile(c, "categories")
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot parse categories file")
	}

	var data []categoryRequestData
	err = sonic.Unmarshal(dataBytes, &data)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	for _, categoryData := range data {
		errStr := d.validateCategoryData(&categoryData)
		if errStr != "" {
			return jsonError(c, fiber.StatusBadRequest, errStr)
		}
	}

	categories, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		categories := make([]uuid.UUID, 0, len(data))

		for _, categoryData := range data {
			category := categoryFromRequestData(&categoryData)

			categoryID, err := d.db.Category().Upsert(ctx, category, override == "true")
			if err != nil {
				if errors.Is(err, store.ErrDuplicateKey) {
					return nil, errors.New(categoryExistsMessage(category))
				}

				return nil, fmt.Errorf("Cannot create category \"%s %s\"", category.Identifier, category.Name)
			}
			categories = append(categories, categoryID)
		}

		return categories, nil
	})
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":      "Categories created",
		"category_ids": categories.([]uuid.UUID),
	})
}

func (d *Driver) categoryFromParam(ctx context.Context, categoryParam string) (*model.Category, string) {
	return fromParam(ctx, categoryParam, "Category", d.db.Category().GetByID)
}

func categoryFromRequestData(data *categoryRequestData) *model.Category {
	return &model.Category{
		Identifier:         data.Identifier,
		Name:               data.Name,
		Subcategory:        data.Subcategory,
		GenericDescription: data.GenericDescription,
		GenericRemediation: data.GenericRemediation,
		LanguagesOrder:     data.LanguagesOrder,
		References:         data.References,
		Source:             data.Source,
	}
}

func categoryExistsMessage(category *model.Category) string {
	subcategory := ""
	if category.Subcategory != "" {
		subcategory = fmt.Sprintf(" (%s)", category.Subcategory)
	}
	return fmt.Sprintf("Category \"%s %s%s\" already exists", category.Identifier, category.Name, subcategory)
}

func (d *Driver) validateCategoryData(category *categoryRequestData) string {
	if category.Identifier == "" {
		return "Identifier is required"
	}

	if category.Name == "" {
		return "Name is required"
	}

	return ""
}
