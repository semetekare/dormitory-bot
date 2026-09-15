package service

import "github.com/dormitory-bot/internal/domain"

func MergePermissions(roles []domain.Role, resourceType, operation string) ([]domain.PermissionContext, bool) {
	allContexts := make(map[domain.PermissionContext]bool)
	hasAnyPermission := false

	for _, role := range roles {
		contexts, allowed := role.HasOperation(resourceType, operation)
		if allowed {
			hasAnyPermission = true
			for _, ctx := range contexts {
				allContexts[ctx] = true
			}
		}
	}

	if !hasAnyPermission {
		return nil, false
	}

	result := make([]domain.PermissionContext, 0, len(allContexts))
	for ctx := range allContexts {
		result = append(result, ctx)
	}

	return result, true
}
