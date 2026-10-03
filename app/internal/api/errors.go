package api

import (
	"github.com/gofiber/fiber/v2"
)

func jsonError(c *fiber.Ctx, status int, message string) error {
	c.Status(status)
	return c.JSON(fiber.Map{"error": message})
}
