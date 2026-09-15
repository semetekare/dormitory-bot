package service

import (
	"github.com/dormitory-bot/internal/domain"
	"github.com/google/uuid"
)

type Guards struct {
	svc *DBService
}

func NewGuards(svc *DBService) *Guards {
	return &Guards{svc: svc}
}

func (g *Guards) HasModuleAccess(dormitoryID, employeeID uuid.UUID, module string) bool {
	return g.svc.rbac.HasModuleAccess(dormitoryID, employeeID, module)
}

func (g *Guards) HasModuleAccessForUser(dormitoryID, userID uuid.UUID, module string) bool {
	return g.svc.rbac.HasModuleAccessForUser(dormitoryID, userID, module)
}

func (g *Guards) CheckPermission(dormitoryID, employeeID uuid.UUID, resource, operation string) bool {
	return g.svc.rbac.CheckPermission(dormitoryID, employeeID, resource, operation)
}

func (g *Guards) CheckPermissionForUser(dormitoryID, userID uuid.UUID, resource, operation string) bool {
	return g.svc.rbac.CheckUserPermission(dormitoryID, userID, resource, operation)
}

type ModuleAccess struct {
	ModuleKey string
	HasAccess bool
}

func (g *Guards) GetModuleAccessMap(dormitoryID, employeeID uuid.UUID) map[string]bool {
	result := make(map[string]bool)
	for _, m := range domain.ResourceRegistry {
		result[m.Key] = g.svc.rbac.HasModuleAccess(dormitoryID, employeeID, m.Key)
	}
	return result
}

func (g *Guards) GetModuleAccessMapForUser(dormitoryID, userID uuid.UUID) map[string]bool {
	result := make(map[string]bool)
	for _, m := range domain.ResourceRegistry {
		result[m.Key] = g.svc.rbac.HasModuleAccessForUser(dormitoryID, userID, m.Key)
	}
	return result
}
