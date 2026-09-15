package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/keyboards"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/templates"
	"github.com/dormitory-bot/internal/types"
	"github.com/google/uuid"
)

func handleStaffManagement(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "read"); !ok {
		return blocked, nil
	}

	if page == 0 {
		userCtx.PushState()
		userCtx.State = fsm.StateStaffManagement
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	staff, err := svc.GetStaffByDormitory(did)
	if err != nil || len(staff) == 0 {
		return types.ResponseWithButtons("👥 В этом общежитии пока нет назначенных сотрудников.\n\nИспользуйте кнопку «Добавить» для назначения ролей.", [][]keyboards.Button{
			{{Text: "➕ Добавить", Payload: "staff:add"}},
			{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
		}), nil
	}

	// Pre-load all roles so custom role names get their DisplayName
	rolenames := buildRoleDisplayLookup(svc)

	msg := "👥 Сотрудники и роли:\n\n"
	items := make([]keyboards.PageableButton, 0, len(staff))
	for _, s := range staff {
		fullName := fmt.Sprintf("%s %s", s.UserLastName, s.UserFirstName)
		if s.UserMiddleName != "" {
			fullName = fmt.Sprintf("%s %s %s", s.UserLastName, s.UserFirstName, s.UserMiddleName)
		}
		personType := "👔"
		if s.PersonType == "student" {
			personType = "🏠"
		}
		roleDisplay := service.MapRoleDisplayWithLookup(s.Role, rolenames)
		msg += fmt.Sprintf("%s %s — %s (%s)\n", personType, fullName, roleDisplay, s.PersonType)
		items = append(items, keyboards.PageableButton{
			Text:    fmt.Sprintf("%s %s — %s", personType, fullName, roleDisplay),
			Payload: fmt.Sprintf("staff:view:%s", s.ID.String()),
		})
	}

	kb := keyboards.BuildPaginatedKeyboard(items, "staff:mgmt", page, 8)

	if hasPermission(userCtx, svc, "staff", "create") {
		kb[len(kb)-1] = append(kb[len(kb)-1], keyboards.Button{Text: "➕ Добавить", Payload: "staff:add"})
	}

	return types.ResponseWithButtons(msg, kb), nil
}

// buildRoleDisplayLookup loads all roles and returns a name→DisplayName map.
func buildRoleDisplayLookup(svc *service.DBService) map[string]string {
	roles, err := svc.ListRoles()
	if err != nil {
		return nil
	}
	m := make(map[string]string, len(roles))
	for _, r := range roles {
		m[r.Name] = r.DisplayName
	}
	return m
}

func handleStaffView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "read"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "staff:view:")
	id, _ := uuid.Parse(roleID)

	did, _ := uuid.Parse(userCtx.DormitoryID)
	staff, err := svc.GetStaffByDormitory(did)
	if err != nil {
		return types.ResponseWithButtons("👥 Сотрудник не найден.", keyboards.GetBackKeyboard()), nil
	}

	var target *domain.UserDormitoryRoleWithUser
	for _, s := range staff {
		if s.ID == id {
			target = &s
			break
		}
	}
	if target == nil {
		return types.ResponseWithButtons("👥 Сотрудник не найден.", keyboards.GetBackKeyboard()), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateStaffView
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	fullName := fmt.Sprintf("%s %s", target.UserLastName, target.UserFirstName)
	if target.UserMiddleName != "" {
		fullName = fmt.Sprintf("%s %s %s", target.UserLastName, target.UserFirstName, target.UserMiddleName)
	}

	srcLabel := ""
	if target.Source == "edr" {
		srcLabel = "\n🔒 Назначено через ЕИС"
	}
	rolenames := buildRoleDisplayLookup(svc)
	roleDisplay := service.MapRoleDisplayWithLookup(target.Role, rolenames)
	msg := fmt.Sprintf("👤 %s\n\n📱 %s\n🔑 Роль: %s\n👤 Тип: %s%s",
		fullName, target.UserPhone, roleDisplay, target.PersonType, srcLabel)
	if target.PlatformUserID != "" {
		if mention := keyboards.FormatMAXMention(target.PlatformUserID, fullName); mention != "" {
			msg += fmt.Sprintf("\n💬 %s", mention)
		} else {
			msg += "\n⚠️ Этот пользователь ещё не заходил в бота"
		}
	}

	buttons := [][]keyboards.Button{}
	if hasPermission(userCtx, svc, "staff", "update") {
		buttons = append(buttons, []keyboards.Button{{
			Text:    "🔄 Сменить роль",
			Payload: fmt.Sprintf("staff:assign_role:%s", target.UserID.String()),
		}})
	}
	if target.Source != "edr" && hasPermission(userCtx, svc, "staff", "delete") {
		buttons = append(buttons, []keyboards.Button{{
			Text:    "🗑️ Снять роль",
			Payload: fmt.Sprintf("staff:remove:%s", target.ID.String()),
		}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleStaffRoleAssign(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "update"); !ok {
		return blocked, nil
	}

	targetUserID := strings.TrimPrefix(payload, "staff:assign_role:")

	userCtx.PushState()
	userCtx.State = fsm.StateStaffRoleAssign
	userCtx.ContextData = map[string]interface{}{
		"target_user_id": targetUserID,
	}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	roles, err := svc.ListRoles()
	if err != nil || len(roles) == 0 {
		return types.ResponseWithButtons("🔑 Нет доступных ролей.", keyboards.GetBackKeyboard()), nil
	}

	targetUID, _ := uuid.Parse(targetUserID)
	targetUser, _ := svc.GetUserByUserID(targetUID)
	personType := ""
	if targetUser != nil {
		personType = targetUser.PersonType
	}

	var filteredRoles []domain.Role
	for _, r := range roles {
		if personType != "" && !r.CanAssignTo(personType) {
			continue
		}
		filteredRoles = append(filteredRoles, r)
	}
	if len(filteredRoles) == 0 {
		return types.ResponseWithButtons(fmt.Sprintf("🔑 Нет ролей, доступных для %s.", personType), keyboards.GetBackKeyboard()), nil
	}

	msg := "🔑 Выберите роль:"
	if targetUser != nil {
		msg = fmt.Sprintf("🔑 Выберите роль для %s %s:", targetUser.LastName, targetUser.FirstName)
	}

	buttons := [][]keyboards.Button{}
	for _, r := range filteredRoles {
		systemLabel := ""
		if r.IsSystem {
			systemLabel = " [системная]"
		}
		buttons = append(buttons, []keyboards.Button{{
			Text:    fmt.Sprintf("%s%s", r.DisplayName, systemLabel),
			Payload: fmt.Sprintf("staff:confirm_role:%s:%s", targetUserID, r.ID.String()),
		}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleStaffConfirmRole(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "update"); !ok {
		return blocked, nil
	}

	payload = strings.TrimPrefix(payload, "staff:confirm_role:")
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		return types.ResponseWithButtons("❌ Ошибка выбора роли.", keyboards.GetBackKeyboard()), nil
	}
	targetUserID, roleID := parts[0], parts[1]

	uid, _ := uuid.Parse(targetUserID)
	rid, _ := uuid.Parse(roleID)
	did, _ := uuid.Parse(userCtx.DormitoryID)
	eid, _ := uuid.Parse(userCtx.EmployeeUUID)

	role, err := svc.GetRoleByID(rid)
	if err != nil {
		return types.ResponseWithButtons("❌ Роль не найдена.", keyboards.GetBackKeyboard()), nil
	}

	err = svc.AssignRoleToUser(uid, did, eid, role.Name)
	if err != nil {
		log.Printf("Error assigning role: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	userCtx.ContextData = nil
	userCtx.ClearNavigationStack()
	userCtx.State = fsm.StateRoleVerified
	userCtx.PushState()
	userCtx.State = fsm.StateStaffManagement
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	targetUser, _ := svc.GetUserByUserID(uid)
	notifyMsg := fmt.Sprintf("✅ Роль «%s» назначена.", role.DisplayName)
	if targetUser != nil {
		notifyMsg = fmt.Sprintf("✅ %s %s назначен роль «%s».",
			targetUser.LastName, targetUser.FirstName, role.DisplayName)
	}

	return types.ResponseWithButtons(notifyMsg, keyboards.GetBackKeyboard()), nil
}

func handleStaffRemove(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "delete"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "staff:remove:")

	return types.ResponseWithButtons("❗️ Точно снять эту роль?",
		[][]keyboards.Button{
			{{Text: "✅ Да, снять", Payload: fmt.Sprintf("staff:confirm_remove:%s", roleID)}},
			{{Text: "❌ Отмена", Payload: keyboards.PayloadNavBack}},
		}), nil
}

func handleStaffRemoveConfirm(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "delete"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "staff:confirm_remove:")
	id, _ := uuid.Parse(roleID)
	did, _ := uuid.Parse(userCtx.DormitoryID)

	if err := svc.RemoveUserDormitoryRole(id, did); err != nil {
		log.Printf("Error removing role: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	userCtx.ContextData = nil
	userCtx.ClearNavigationStack()
	userCtx.State = fsm.StateRoleVerified
	userCtx.PushState()
	userCtx.State = fsm.StateStaffManagement
	fsmMgr.Set(ctx, userCtx)

	return types.ResponseWithButtons("✅ Роль снята.", keyboards.GetBackKeyboard()), nil
}

func handleStaffAdd(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "staff", "create"); !ok {
		return blocked, nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingStaffData
	userCtx.ContextData = map[string]interface{}{"step": "phone"}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons("Введите номер телефона пользователя для добавления (10 цифр, например: 9123456789):", keyboards.GetBackKeyboard()), nil
}

func HandleStaffMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingStaffData {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)
	did, _ := uuid.Parse(userCtx.DormitoryID)
	eid, _ := uuid.Parse(userCtx.EmployeeUUID)

	switch step {
	case "phone":
		phone := strings.TrimSpace(msg)
		if len(phone) == 10 {
			phone = "7" + phone
		}

		user, err := svc.GetUserByPhone(phone)
		if err != nil || user == nil {
			return types.ResponseWithButtons("❌ Пользователь с таким телефоном не найден.", keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData["target_user_id"] = user.ID.String()
		userCtx.ContextData["target_name"] = fmt.Sprintf("%s %s", user.LastName, user.FirstName)
		userCtx.ContextData["step"] = "role"

		roles, _ := svc.ListRoles()

		var filteredRoles []domain.Role
		for _, r := range roles {
			if !r.CanAssignTo(user.PersonType) {
				continue
			}
			filteredRoles = append(filteredRoles, r)
		}
		if len(filteredRoles) == 0 {
			return types.ResponseWithButtons(fmt.Sprintf("🔑 Нет ролей, доступных для типа «%s».", user.PersonType), keyboards.GetBackKeyboard()), nil
		}

		msg := fmt.Sprintf("👤 %s %s\n\n🔑 Выберите роль:", user.LastName, user.FirstName)
		buttons := [][]keyboards.Button{}
		for _, r := range filteredRoles {
			systemLabel := ""
			if r.IsSystem {
				systemLabel = " [системная]"
			}
			buttons = append(buttons, []keyboards.Button{{
				Text:    fmt.Sprintf("%s%s", r.DisplayName, systemLabel),
				Payload: fmt.Sprintf("staff:add_role:%s", r.ID.String()),
			}})
		}
		buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Отмена", Payload: keyboards.PayloadNavBack}})

		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons(msg, buttons), nil

	case "role":
		rid, _ := uuid.Parse(msg)
		role, err := svc.GetRoleByID(rid)
		if err != nil {
			return types.ResponseWithButtons("❌ Роль не найдена.", keyboards.GetBackKeyboard()), nil
		}

		targetUserID, _ := userCtx.ContextData["target_user_id"].(string)
		uid, _ := uuid.Parse(targetUserID)

		if err := svc.AssignRoleToUser(uid, did, eid, role.Name); err != nil {
			log.Printf("Error assigning role (staff add): %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateStaffManagement
		fsmMgr.Set(ctx, userCtx)

		return types.ResponseWithButtons(
			fmt.Sprintf("✅ Роль «%s» назначена.", role.DisplayName),
			keyboards.GetBackKeyboard(),
		), nil
	}

	return nil, nil
}

func HandleStaffCallback(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if keyboards.HasPageSuffix(payload) {
		base := keyboards.StripPageSuffix(payload)
		page := keyboards.ParsePage(payload)
		if strings.HasPrefix(base, "staff:mgmt") {
			return handleStaffManagement(ctx, userCtx, svc, fsmMgr, page)
		}
	}

	switch {
	case strings.HasPrefix(payload, "staff:view:"):
		return handleStaffView(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "staff:assign_role:"):
		return handleStaffRoleAssign(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "staff:confirm_role:"):
		return handleStaffConfirmRole(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "staff:confirm_remove:"):
		return handleStaffRemoveConfirm(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "staff:remove:"):
		return handleStaffRemove(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "staff:add":
		return handleStaffAdd(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "staff:add_role:"):
		rid := strings.TrimPrefix(payload, "staff:add_role:")
		return HandleStaffMessage(ctx, userCtx, rid, svc, fsmMgr)
	}

	return nil, nil
}
