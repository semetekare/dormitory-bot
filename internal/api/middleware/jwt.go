package middleware

import (
	"strings"

	"github.com/dormitory-bot/internal/auth"
	"github.com/dormitory-bot/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// JWTAuthMiddleware проверяет Bearer токен, извлекает claims в c.Locals.
func JWTAuthMiddleware(jwtSvc *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid Authorization format, expected: Bearer <token>")
		}

		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("person_type", claims.PersonType)
		if claims.DormitoryID != "" {
			c.Locals("dormitory_id", claims.DormitoryID)
		}
		if claims.EmployeeID != "" {
			c.Locals("employee_id", claims.EmployeeID)
		}
		return c.Next()
	}
}

// RequirePermissionJWT проверяет права из JWT claims (без X-Employee-ID заголовков).
// Читает employee_id и dormitory_id из c.Locals (установлены JWTAuthMiddleware).
func RequirePermissionJWT(svc *service.DBService, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		employeeIDStr, _ := c.Locals("employee_id").(string)
		dormitoryIDStr, _ := c.Locals("dormitory_id").(string)

		if employeeIDStr == "" || dormitoryIDStr == "" {
			return fiber.NewError(fiber.StatusForbidden, "employee access required — no employee_id or dormitory_id in token")
		}

		employeeID, err1 := uuid.Parse(employeeIDStr)
		dormitoryID, err2 := uuid.Parse(dormitoryIDStr)
		if err1 != nil || err2 != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid UUID in token claims")
		}

		if !svc.CheckPermission(dormitoryID, employeeID, resource, action) {
			return fiber.NewError(fiber.StatusForbidden, "permission denied")
		}
		return c.Next()
	}
}
