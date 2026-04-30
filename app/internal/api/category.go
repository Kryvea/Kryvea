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
	// parse request body
	data := &categoryRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr := d.validateCategoryData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	category := &model.Category{
		Identifier:         data.Identifier,
		Name:               data.Name,
		Subcategory:        data.Subcategory,
		GenericDescription: data.GenericDescription,
		GenericRemediation: data.GenericRemediation,
		LanguagesOrder:     data.LanguagesOrder,
		References:         data.References,
		Source:             data.Source,
	}

	// insert category into database
	categoryID, err := d.db.Category().Insert(c.UserContext(), category)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if errors.Is(err, store.ErrDuplicateKey) {
			subcategory := ""
			if category.Subcategory != "" {
				subcategory = fmt.Sprintf(" (%s)", category.Subcategory)
			}
			return c.JSON(fiber.Map{
				"error": fmt.Sprintf("Category \"%s %s%s\" already exists", category.Identifier, category.Name, subcategory),
			})
		}

		return c.JSON(fiber.Map{
			"error": "Cannot create category",
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Category created",
		"category_id": categoryID,
	})
}

func (d *Driver) UpdateCategory(c *fiber.Ctx) error {
	// parse category param
	category, errStr := d.categoryFromParam(c.UserContext(), c.Params("category"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// parse request body
	data := &categoryRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr = d.validateCategoryData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	newCategory := &model.Category{
		Identifier:         data.Identifier,
		Name:               data.Name,
		Subcategory:        data.Subcategory,
		GenericDescription: data.GenericDescription,
		GenericRemediation: data.GenericRemediation,
		LanguagesOrder:     data.LanguagesOrder,
		References:         data.References,
		Source:             data.Source,
	}

	// update category in database
	err := d.db.Category().Update(c.UserContext(), category.ID, newCategory)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if errors.Is(err, store.ErrDuplicateKey) {
			subcategory := ""
			if newCategory.Subcategory != "" {
				subcategory = fmt.Sprintf(" (%s)", newCategory.Subcategory)
			}
			return c.JSON(fiber.Map{
				"error": fmt.Sprintf("Category \"%s %s%s\" already exists", newCategory.Identifier, newCategory.Name, subcategory),
			})
		}

		return c.JSON(fiber.Map{
			"error": "Cannot update category",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Category updated",
	})
}

func (d *Driver) DeleteCategory(c *fiber.Ctx) error {
	// parse category param
	category, errStr := d.categoryFromParam(c.UserContext(), c.Params("category"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	if err := d.db.Category().Delete(c.UserContext(), category.ID); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot delete category",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Category deleted",
	})
}

func (d *Driver) SearchCategories(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Query is required",
		})
	}

	categories, err := d.db.Category().Search(c.UserContext(), query)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot search categories",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(categories)
}

func (d *Driver) GetCategories(c *fiber.Ctx) error {
	categories, err := d.db.Category().GetAll(c.UserContext())
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get categories",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(categories)
}

func (d *Driver) ExportCategories(c *fiber.Ctx) error {
	categories, err := d.db.Category().GetAll(c.UserContext())
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get categories",
		})
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Disposition", "attachment; filename=categories.json")
	return c.JSON(categories)
}

func (d *Driver) GetCategory(c *fiber.Ctx) error {
	// parse category param
	category, errStr := d.categoryFromParam(c.UserContext(), c.Params("category"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(category)
}

func (d *Driver) UploadCategories(c *fiber.Ctx) error {
	// parse override parameter
	override := c.FormValue("override")

	// parse request body
	dataBytes, err := util.ParseFormFile(c, "categories")
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot parse categories file",
		})
	}

	var data []categoryRequestData
	err = sonic.Unmarshal(dataBytes, &data)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate each category data
	for _, categoryData := range data {
		errStr := d.validateCategoryData(&categoryData)
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": errStr,
			})
		}
	}

	categories, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		categories := make([]uuid.UUID, 0, len(data))

		// insert each category into database
		for _, categoryData := range data {
			category := &model.Category{
				Identifier:         categoryData.Identifier,
				Name:               categoryData.Name,
				Subcategory:        categoryData.Subcategory,
				GenericDescription: categoryData.GenericDescription,
				GenericRemediation: categoryData.GenericRemediation,
				LanguagesOrder:     categoryData.LanguagesOrder,
				References:         categoryData.References,
				Source:             categoryData.Source,
			}

			categoryID, err := d.db.Category().Upsert(ctx, category, override == "true")
			if err != nil {
				if errors.Is(err, store.ErrDuplicateKey) {
					subcategory := ""
					if category.Subcategory != "" {
						subcategory = fmt.Sprintf(" (%s)", category.Subcategory)
					}
					return nil, fmt.Errorf("Category \"%s %s%s\" already exists", category.Identifier, category.Name, subcategory)
				}

				return nil, fmt.Errorf("Cannot create category \"%s %s\"", category.Identifier, category.Name)
			}
			categories = append(categories, categoryID)
		}

		return categories, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":      "Categories created",
		"category_ids": categories.([]uuid.UUID),
	})
}

func (d *Driver) categoryFromParam(ctx context.Context, categoryParam string) (*model.Category, string) {
	if categoryParam == "" {
		return nil, "Category ID is required"
	}

	categoryID, err := util.ParseUUID(categoryParam)
	if err != nil {
		return nil, "Invalid category ID"
	}

	category, err := d.db.Category().GetByID(ctx, categoryID)
	if err != nil {
		return nil, "Invalid category ID"
	}

	return category, ""
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
