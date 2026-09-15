package middleware

import (
	"github.com/dormitory-bot/internal/service"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(svc *service.DBService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		if userID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing X-User-ID header")
		}
		c.Locals("user_id", userID)
		return c.Next()
	}
}
