package api

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
)

const bytesPerMB = 1024 * 1024

// MaxImageSize is expressed as float64 MB
type settingRequestData struct {
	MaxImageSize            float64                    `json:"max_image_size"`
	DefaultCategoryLanguage string                     `json:"default_category_language"`
	VulnerabilityFields     []model.VulnerabilityField `json:"vulnerability_fields"`
}

type settingResponseData struct {
	ID                      string                     `json:"id"`
	MaxImageSize            float64                    `json:"max_image_size"`
	DefaultCategoryLanguage string                     `json:"default_category_language"`
	VulnerabilityFields     []model.VulnerabilityField `json:"vulnerability_fields"`
}

func (d *Driver) GetSettings(c *fiber.Ctx) error {
	settings, err := d.db.Setting().Get(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot retrieve settings")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(settingResponseData{
		ID:                      settings.ID.String(),
		MaxImageSize:            float64(settings.MaxImageSize) / bytesPerMB,
		DefaultCategoryLanguage: settings.DefaultCategoryLanguage,
		VulnerabilityFields:     settings.VulnerabilityFields,
	})
}

func (d *Driver) UpdateSettings(c *fiber.Ctx) error {
	data := &settingRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr := d.validateSettingData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	setting := &model.Setting{
		MaxImageSize:            int64(data.MaxImageSize * bytesPerMB),
		DefaultCategoryLanguage: data.DefaultCategoryLanguage,
		VulnerabilityFields:     data.VulnerabilityFields,
	}

	_, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		return nil, d.db.Setting().Update(ctx, setting)
	})
	if err != nil {
		if errors.Is(err, model.ErrInvalidVulnerabilityFields) {
			return jsonError(c, fiber.StatusBadRequest, err.Error())
		}
		return jsonError(c, fiber.StatusBadRequest, "Cannot update settings")
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message": "Settings updated",
	})
}

func (d *Driver) validateSettingData(setting *settingRequestData) string {
	if _, err := language.Parse(setting.DefaultCategoryLanguage); err != nil {
		return "invalid language"
	}

	return ""
}

func (d *Driver) GetVulnerabilityFields(c *fiber.Ctx) error {
	settings, err := d.db.Setting().Get(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve vulnerability custom fields")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(settings.VulnerabilityFields)
}

func (d *Driver) GetVulnerabilityFieldImpact(c *fiber.Ctx) error {
	impact, err := d.db.Setting().GetVulnerabilityFieldImpact(c.UserContext(), c.Params("key"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return jsonError(c, fiber.StatusNotFound, "Custom field not found")
		}
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve custom field deletion impact")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(impact)
}
