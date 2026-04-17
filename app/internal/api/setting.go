package api

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
)

// MaxImageSize is expressed as float64 MB
// Collective wisdom dictated this
// I, however, remain in dignified dissent
type settingRequestData struct {
	MaxImageSize            float64 `json:"max_image_size" bson:"max_image_size"`
	DefaultCategoryLanguage string  `json:"default_category_language" bson:"default_category_language"`
}

func (d *Driver) GetSettings(c *fiber.Ctx) error {
	settings, err := d.mongo.Setting().Get(context.Background())
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve settings",
		})
	}

	settings.MaxImageSize /= 1024 * 1024

	c.Status(fiber.StatusOK)
	return c.JSON(settings)
}

func (d *Driver) UpdateSettings(c *fiber.Ctx) error {
	// parse request body
	data := &settingRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr := d.validateSettingData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// Image size is received as Mbytes, convert to bytes
	maxImageSize := int64(data.MaxImageSize * 1024 * 1024)

	setting := &mongo.Setting{
		MaxImageSize:            maxImageSize,
		MaxImageSizeMB:          data.MaxImageSize,
		DefaultCategoryLanguage: data.DefaultCategoryLanguage,
	}

	// insert update into database
	err := d.mongo.Setting().Update(context.Background(), setting)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot update settings",
		})
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
