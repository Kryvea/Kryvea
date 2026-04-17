package api

import (
	"context"
	"strings"
	"time"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/fiber/v2"
)

func (d *Driver) SessionMiddleware(c *fiber.Ctx) error {
	if c.Path() == "/api/login" && c.Method() == fiber.MethodPost {
		return c.Next()
	}

	session := c.Cookies("kryvea")
	token, err := crypto.ParseToken(session)
	if err != nil {
		util.ClearCookies(c)
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	user, err := d.mongo.User().GetByToken(context.Background(), token)
	if err != nil || user.TokenExpiry.Before(time.Now()) || (!user.DisabledAt.IsZero() && user.DisabledAt.Before(time.Now())) {
		util.ClearCookies(c)
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if user.PasswordExpiry.Before(time.Now()) {
		if c.Path() == "/api/password/reset" && c.Method() == fiber.MethodPost {
			c.Locals("user", user)
			return c.Next()
		}

		c.Status(fiber.StatusUnauthorized)
		util.ClearCookies(c)
		return c.JSON(fiber.Map{
			"error": "Password expired",
		})
	}

	if time.Until(user.TokenExpiry) < mongo.TokenRefreshThreshold {
		err := d.mongo.User().RefreshUserToken(context.Background(), user)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(fiber.Map{
				"error": "Failed to refresh session",
			})
		}
		util.SetSessionCookies(c, user.Role, user.Token, user.TokenExpiry)
	}

	c.Locals("user", user)

	return c.Next()
}

func (d *Driver) AdminMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	if user.Role != mongo.RoleAdmin {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	return c.Next()
}

func (d *Driver) ContentTypeMiddleware(c *fiber.Ctx) error {
	method := c.Method()
	path := c.Path()
	contentType := c.Get(fiber.HeaderContentType)

	if strings.HasSuffix(path, "/upload") && method == fiber.MethodPost {
		if !strings.HasPrefix(contentType, fiber.MIMEMultipartForm) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Content-Type must be multipart/form-data",
			})
		}
		return c.Next()
	}

	if (method == fiber.MethodPost || method == fiber.MethodPatch) &&
		c.Request().Header.ContentLength() > 0 &&
		!strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content-Type must be application/json",
		})
	}

	return c.Next()
}
