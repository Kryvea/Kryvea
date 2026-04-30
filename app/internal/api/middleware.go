package api

import (
	"time"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/Kryvea/Kryvea/internal/model"
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

	user, err := d.db.User().GetByToken(c.UserContext(), token)
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

	if time.Until(user.TokenExpiry) < model.TokenRefreshThreshold {
		err := d.db.User().RefreshUserToken(c.UserContext(), user)
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
	user := c.Locals("user").(*model.User)

	if user.Role != model.RoleAdmin {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	return c.Next()
}
