package api

import (
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
)

const bytesPerMB = 1024 * 1024

// MaxImageSize is expressed as float64 MB
type settingRequestData struct {
	MaxImageSize            float64 `json:"max_image_size"`
	DefaultCategoryLanguage string  `json:"default_category_language"`
}

type settingResponseData struct {
	ID                      string  `json:"id"`
	MaxImageSize            float64 `json:"max_image_size"`
	DefaultCategoryLanguage string  `json:"default_category_language"`
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
	}

	err := d.db.Setting().Update(c.UserContext(), setting)
	if err != nil {
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
