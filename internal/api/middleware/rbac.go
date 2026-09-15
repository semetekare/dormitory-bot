package middleware

import (
	"github.com/dormitory-bot/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequirePermission(svc *service.DBService, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		employeeIDStr := c.Get("X-Employee-ID")
		dormitoryIDStr := c.Get("X-Dormitory-ID")

		if employeeIDStr == "" || dormitoryIDStr == "" {
			return fiber.NewError(fiber.StatusForbidden, "missing X-Employee-ID or X-Dormitory-ID header")
		}

		employeeID, err1 := uuid.Parse(employeeIDStr)
		dormitoryID, err2 := uuid.Parse(dormitoryIDStr)

		if err1 != nil || err2 != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid UUID")
		}

		if !svc.CheckPermission(dormitoryID, employeeID, resource, action) {
			return fiber.NewError(fiber.StatusForbidden, "permission denied")
		}

		return c.Next()
	}
}
