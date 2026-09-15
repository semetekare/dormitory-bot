package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/keyboards"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/types"
	"github.com/google/uuid"
)

func guardPermission(userCtx *fsm.UserContext, svc *service.DBService, resource, operation string) (*types.MessageResponse, bool) {
	did, _ := uuid.Parse(userCtx.DormitoryID)

	if userCtx.EmployeeUUID != "" {
		eid, _ := uuid.Parse(userCtx.EmployeeUUID)
		if svc.Guards.CheckPermission(did, eid, resource, operation) {
			return nil, true
		}
	} else if userCtx.UserUUID != "" {
		uid, _ := uuid.Parse(userCtx.UserUUID)
		if svc.Guards.CheckPermissionForUser(did, uid, resource, operation) {
			return nil, true
		}
	}

	log.Printf("Permission denied: user=%d resource=%s op=%s dormitory=%s role=%s",
		userCtx.UserID, resource, operation, userCtx.DormitoryID, userCtx.EmployeeRole)

	blocked := types.ResponseWithButtons("❌ Недостаточно прав.", keyboards.GetBackKeyboard())
	return blocked, false
}

func hasPermission(userCtx *fsm.UserContext, svc *service.DBService, resource, operation string) bool {
	did, _ := uuid.Parse(userCtx.DormitoryID)
	if userCtx.EmployeeUUID != "" {
		eid, _ := uuid.Parse(userCtx.EmployeeUUID)
		return svc.Guards.CheckPermission(did, eid, resource, operation)
	}
	if userCtx.UserUUID != "" {
		uid, _ := uuid.Parse(userCtx.UserUUID)
		return svc.Guards.CheckPermissionForUser(did, uid, resource, operation)
	}
	return false
}

type adminMenuBuilder struct {
	svc          *service.DBService
	dormitoryID  uuid.UUID
	employeeUUID string
	userUUID     string
	roleDisplay  string
}

func newAdminMenuBuilderFromContext(svc *service.DBService, userCtx *fsm.UserContext) *adminMenuBuilder {
	return NewAdminMenuBuilderFromContext(svc, userCtx)
}

func NewAdminMenuBuilderFromContext(svc *service.DBService, userCtx *fsm.UserContext) *adminMenuBuilder {
	did, _ := uuid.Parse(userCtx.DormitoryID)
	rolenames := buildRoleDisplayLookup(svc)
	roleDisplay := service.MapRoleDisplayWithLookup(userCtx.EmployeeRole, rolenames)

	if userCtx.EmployeeUUID != "" {
		actualRole, err := svc.GetEmployeeRoleInDormitory(userCtx.EmployeeUUID, userCtx.DormitoryID)
		if err == nil && actualRole != "" {
			roleDisplay = service.MapRoleDisplayWithLookup(actualRole, rolenames)
			if actualRole != userCtx.EmployeeRole {
				log.Printf("newAdminMenuBuilder: syncing stale role for employee %s: %s -> %s",
					userCtx.EmployeeUUID, userCtx.EmployeeRole, actualRole)
				userCtx.EmployeeRole = actualRole
			}
		}
	} else if userCtx.UserUUID != "" && userCtx.EmployeeRole != "" {
		uid, _ := uuid.Parse(userCtx.UserUUID)
		roles, err := svc.GetUserRolesForDormitory(uid, did)
		if err == nil && len(roles) > 0 {
			actualRole := string(roles[0].Role)
			rolenames := buildRoleDisplayLookup(svc)
			roleDisplay = service.MapRoleDisplayWithLookup(actualRole, rolenames)
			if actualRole != userCtx.EmployeeRole {
				log.Printf("newAdminMenuBuilder: syncing stale role for resident %s: %s -> %s",
					userCtx.UserUUID, userCtx.EmployeeRole, actualRole)
				userCtx.EmployeeRole = actualRole
			}
		} else if len(roles) == 0 {
			roleDisplay = ""
		}
	}

	return &adminMenuBuilder{
		svc:          svc,
		dormitoryID:  did,
		employeeUUID: userCtx.EmployeeUUID,
		userUUID:     userCtx.UserUUID,
		roleDisplay:  roleDisplay,
	}
}

func (b *adminMenuBuilder) RoleDisplay() string {
	return b.roleDisplay
}

func (b *adminMenuBuilder) build(ctx context.Context) [][]keyboards.Button {
	return b.Build(ctx)
}

func (b *adminMenuBuilder) Build(ctx context.Context) [][]keyboards.Button {
	var rows [][]keyboards.Button

	if b.employeeUUID != "" {
		eid, _ := uuid.Parse(b.employeeUUID)
		b.buildEmployeeMenu(ctx, eid, &rows)
	} else if b.userUUID != "" {
		uid, _ := uuid.Parse(b.userUUID)
		b.buildUserMenu(ctx, uid, &rows)
	}

	return rows
}

func (b *adminMenuBuilder) buildEmployeeMenu(_ context.Context, eid uuid.UUID, rows *[][]keyboards.Button) {
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "module") {
		*rows = append(*rows, []keyboards.Button{{Text: "🤖 Модули ботов", Payload: "menu:modules"}})
	}
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "laundry") {
		*rows = append(*rows, []keyboards.Button{{Text: "👕 Управление прачкой", Payload: "menu:laundry_mgmt"}})
	}
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "room") {
		*rows = append(*rows, []keyboards.Button{{Text: "🚪 Комнаты и жильцы", Payload: "menu:room_mgmt"}})
	}
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "cleaning") {
		*rows = append(*rows, []keyboards.Button{{Text: "🧹 Дежурства", Payload: "menu:cleaning_mgmt"}})
	}
	*rows = append(*rows, []keyboards.Button{{Text: "💬 Ссылки на чаты", Payload: "menu:chat_links_mgmt"}})
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "reference") {
		*rows = append(*rows, []keyboards.Button{{Text: "📖 Справки", Payload: "menu:reference_mgmt"}})
	}
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "role") {
		*rows = append(*rows, []keyboards.Button{{Text: "🔑 Роли", Payload: "menu:role_mgmt"}})
	}
	if b.svc.Guards.HasModuleAccess(b.dormitoryID, eid, "staff") {
		*rows = append(*rows, []keyboards.Button{{Text: "👥 Сотрудники", Payload: "menu:staff_mgmt"}})
	}
}

func (b *adminMenuBuilder) buildUserMenu(_ context.Context, uid uuid.UUID, rows *[][]keyboards.Button) {
	if b.svc.Guards.HasModuleAccessForUser(b.dormitoryID, uid, "laundry") {
		*rows = append(*rows, []keyboards.Button{{Text: "👕 Управление прачкой", Payload: "menu:laundry_mgmt"}})
	}
	if b.svc.Guards.HasModuleAccessForUser(b.dormitoryID, uid, "cleaning") {
		*rows = append(*rows, []keyboards.Button{{Text: "🧹 Дежурства", Payload: "menu:cleaning_mgmt"}})
	}
	if b.svc.Guards.HasModuleAccessForUser(b.dormitoryID, uid, "chat_link") {
		*rows = append(*rows, []keyboards.Button{{Text: "💬 Ссылки на чаты", Payload: "menu:chat_links_mgmt"}})
	}
	*rows = append(*rows, []keyboards.Button{{Text: "◀ Выйти из роли", Payload: "resident_role:exit"}})
}

func buildResidentMainMenu(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService) [][]keyboards.Button {
	kb := keyboards.GetMainMenuKeyboard()

	if userCtx.UserUUID == "" || userCtx.DormitoryID == "" {
		return kb
	}

	uid, err1 := uuid.Parse(userCtx.UserUUID)
	did, err2 := uuid.Parse(userCtx.DormitoryID)
	if err1 != nil || err2 != nil {
		return kb
	}

	count, err := svc.GetUserResidentRoleCount(uid, did)
	if err != nil || count == 0 {
		return kb
	}

	roles, err := svc.GetUserRolesForDormitory(uid, did)
	if err != nil || len(roles) == 0 {
		return kb
	}

	rolenames := buildRoleDisplayLookup(svc)
	var roleNames []string
	for _, r := range roles {
		roleNames = append(roleNames, service.MapRoleDisplayWithLookup(string(r.Role), rolenames))
	}

	roleLabel := "🔑 Роль"
	if len(roleNames) > 0 {
		roleLabel = fmt.Sprintf("🔑 Роль: %s", roleNames[0])
	}

	kb = append(kb, []keyboards.Button{{Text: roleLabel, Payload: "resident_role:select"}})
	return kb
}
