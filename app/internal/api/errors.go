package api

import (
	"github.com/gofiber/fiber/v2"
)

type errorResponse struct {
	Error   string   `json:"error,omitempty"`
	Details string   `json:"details,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

func jsonError(c *fiber.Ctx, status int, message string) error {
	c.Status(status)
	return c.JSON(errorResponse{Error: message})
}
