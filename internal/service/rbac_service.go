package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
)

type RBACService struct {
	roleRepo                  *repository.BaseRepository[domain.Role]
	employeeRepo              *repository.BaseRepository[domain.Employee]
	employeeDormitoryRoleRepo *repository.EmployeeDormitoryRoleRepository
	userDormitoryRoleRepo     *repository.UserDormitoryRoleRepository
	dbService                 *DBService
}

func NewRBACService(
	roleRepo *repository.BaseRepository[domain.Role],
	employeeRepo *repository.BaseRepository[domain.Employee],
	edrRepo *repository.EmployeeDormitoryRoleRepository,
	udrRepo *repository.UserDormitoryRoleRepository,
	dbService *DBService,
) *RBACService {
	return &RBACService{
		roleRepo:                  roleRepo,
		employeeRepo:              employeeRepo,
		employeeDormitoryRoleRepo: edrRepo,
		userDormitoryRoleRepo:     udrRepo,
		dbService:                 dbService,
	}
}

func (s *RBACService) CheckPermission(dormitoryID uuid.UUID, employeeID uuid.UUID, resource string, operation string) bool {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return false
	}

	userID := employee.UserID
	roles, err := s.getEmployeeRoles(employeeID, userID, dormitoryID)
	if err != nil || len(roles) == 0 {
		return false
	}

	_, allowed := MergePermissions(roles, resource, operation)

	fmt.Printf("[SECURITY] Employee %s (%s) via user %s checks %s:%s on dormitory %s - Result: %v\n",
		employee.ID, employee.Role, userID, resource, operation, dormitoryID, allowed)

	return allowed
}

func (s *RBACService) CheckUserPermission(dormitoryID, userID uuid.UUID, resource, operation string) bool {
	roles, err := s.getRolesForUser(userID, dormitoryID)
	if err != nil || len(roles) == 0 {
		return false
	}

	_, allowed := MergePermissions(roles, resource, operation)

	fmt.Printf("[SECURITY] User %s checks %s:%s on dormitory %s - Result: %v\n",
		userID, resource, operation, dormitoryID, allowed)

	return allowed
}

func (s *RBACService) HasModuleAccess(dormitoryID, employeeID uuid.UUID, module string) bool {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return false
	}
	roles, err := s.getEmployeeRoles(employeeID, employee.UserID, dormitoryID)
	if err != nil || len(roles) == 0 {
		return false
	}

	for _, role := range roles {
		for _, op := range []string{"read", "create", "update", "delete"} {
			if _, allowed := role.HasOperation(module, op); allowed {
				return true
			}
		}
		if _, allowed := role.HasOperation(module, "*"); allowed {
			return true
		}
	}
	return false
}

func (s *RBACService) HasModuleAccessForUser(dormitoryID, userID uuid.UUID, module string) bool {
	roles, err := s.getRolesForUser(userID, dormitoryID)
	if err != nil || len(roles) == 0 {
		return false
	}

	for _, role := range roles {
		for _, op := range []string{"read", "create", "update", "delete"} {
			if _, allowed := role.HasOperation(module, op); allowed {
				return true
			}
		}
		if _, allowed := role.HasOperation(module, "*"); allowed {
			return true
		}
	}
	return false
}

func (s *RBACService) GetEffectivePermissions(dormitoryID, employeeID uuid.UUID) (domain.JSONAccessMap, error) {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, err
	}
	roles, err := s.getEmployeeRoles(employeeID, employee.UserID, dormitoryID)
	if err != nil {
		return nil, err
	}
	return s.mergeEffectivePermissions(roles), nil
}

func (s *RBACService) GetEffectivePermissionsForUser(dormitoryID, userID uuid.UUID) (domain.JSONAccessMap, error) {
	roles, err := s.getRolesForUser(userID, dormitoryID)
	if err != nil {
		return nil, err
	}
	return s.mergeEffectivePermissions(roles), nil
}

func (s *RBACService) mergeEffectivePermissions(roles []domain.Role) domain.JSONAccessMap {
	result := make(domain.JSONAccessMap)
	for _, role := range roles {
		for resourceKey, perms := range role.JSONAccess {
			if existing, ok := result[resourceKey]; ok {
				for op, val := range perms {
					existing[op] = val
				}
			} else {
				copied := make(domain.ResourcePermissions)
				for op, val := range perms {
					copied[op] = val
				}
				result[resourceKey] = copied
			}
		}
	}
	return result
}

func (s *RBACService) GetAllowedPermissions(dormitoryID, employeeID uuid.UUID) (domain.JSONAccessMap, error) {
	return s.GetEffectivePermissions(dormitoryID, employeeID)
}

func (s *RBACService) getEmployeeRoles(employeeID, userID, dormitoryID uuid.UUID) ([]domain.Role, error) {
	edrRoles, err := s.employeeDormitoryRoleRepo.GetByEmployeeAndDormitory(employeeID, dormitoryID)
	if err != nil {
		return nil, err
	}

	var roles []domain.Role
	for _, edr := range edrRoles {
		role, err := s.roleRepo.FindByField("name", string(edr.Role))
		if err == nil {
			roles = append(roles, *role)
		}
	}

	udrRoles, _ := s.userDormitoryRoleRepo.GetByUserAndDormitory(userID, dormitoryID)
	for _, udr := range udrRoles {
		role, err := s.roleRepo.FindByField("name", string(udr.Role))
		if err == nil {
			roles = append(roles, *role)
		}
	}

	employee, err := s.employeeRepo.GetByID(employeeID)
	if err == nil {
		role, err := s.roleRepo.FindByField("name", string(employee.Role))
		if err == nil {
			roles = append(roles, *role)
		}
	}

	return roles, nil
}

func (s *RBACService) getRolesForUser(userID, dormitoryID uuid.UUID) ([]domain.Role, error) {
	var roles []domain.Role

	udrRoles, err := s.userDormitoryRoleRepo.GetByUserAndDormitory(userID, dormitoryID)
	if err != nil {
		return nil, err
	}
	for _, udr := range udrRoles {
		role, err := s.roleRepo.FindByField("name", string(udr.Role))
		if err == nil {
			roles = append(roles, *role)
		}
	}

	employee, err := s.employeeRepo.FindByField("user_id", userID)
	if err == nil {
		edrRoles, _ := s.employeeDormitoryRoleRepo.GetByEmployeeAndDormitory(employee.ID, dormitoryID)
		for _, edr := range edrRoles {
			role, err := s.roleRepo.FindByField("name", string(edr.Role))
			if err == nil {
				roles = append(roles, *role)
			}
		}
		role, err := s.roleRepo.FindByField("name", string(employee.Role))
		if err == nil {
			roles = append(roles, *role)
		}
	}

	return roles, nil
}
