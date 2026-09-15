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

var opDisplayNames = map[string]string{
	"read": "Просмотр", "create": "Добавить", "update": "Изменить", "delete": "Удалить",
}

func handleRoleManagement(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "read"); !ok {
		return blocked, nil
	}

	if page == 0 {
		userCtx.PushState()
		userCtx.State = fsm.StateRoleManagement
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	roles, err := svc.ListRoles()
	if err != nil || len(roles) == 0 {
		return types.ResponseWithButtons("🔑 Нет доступных ролей.", keyboards.GetBackKeyboard()), nil
	}

	msg := "🔑 Управление ролями:\n\n💡 Как работает: каждая роль имеет приоритет (чем выше число — тем выше приоритет). Пользователь получает права роли с наивысшим приоритетом. Если у роли нет доступа к модулю — операции запрещены.\n\n"
	items := make([]keyboards.PageableButton, 0, len(roles))
	for _, r := range roles {
		systemLabel := ""
		if r.IsSystem {
			systemLabel = " [системная]"
		}
		msg += fmt.Sprintf("• %s%s (приоритет: %d)\n", r.DisplayName, systemLabel, r.Priority)
		items = append(items, keyboards.PageableButton{
			Text:    fmt.Sprintf("%s%s", r.DisplayName, systemLabel),
			Payload: fmt.Sprintf("role:view:%s", r.ID.String()),
		})
	}

	kb := keyboards.BuildPaginatedKeyboard(items, "role:mgmt", page, 10)

	if hasPermission(userCtx, svc, "role", "create") {
		if len(kb) > 0 {
			kb[len(kb)-1] = append(kb[len(kb)-1], keyboards.Button{Text: "➕ Создать", Payload: "role:create"})
		}
	}

	return types.ResponseWithButtons(msg, kb), nil
}

func handleRoleView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "read"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "role:view:")
	id, _ := uuid.Parse(roleID)

	role, err := svc.GetRoleByID(id)
	if err != nil {
		return types.ResponseWithButtons("🔑 Роль не найдена.", keyboards.GetBackKeyboard()), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateRoleView
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	systemLabel := ""
	if role.IsSystem {
		systemLabel = "\n⚠️ Системная роль (нельзя удалить)"
	}
	msg := fmt.Sprintf("🔑 %s%s\n\nРазрешения:\n", role.DisplayName, systemLabel)
	msg += formatPermissionMatrix(role.JSONAccess)

	buttons := [][]keyboards.Button{}
	if !role.IsSystem && hasPermission(userCtx, svc, "role", "update") {
		buttons = append(buttons, []keyboards.Button{{
			Text:    "✏️ Редактировать",
			Payload: fmt.Sprintf("role:edit:%s", role.ID.String()),
		}})
	}
	if !role.IsSystem && hasPermission(userCtx, svc, "role", "delete") {
		buttons = append(buttons, []keyboards.Button{{
			Text:    "🗑️ Удалить",
			Payload: fmt.Sprintf("role:delete:%s", role.ID.String()),
		}})
	}
	if hasPermission(userCtx, svc, "role", "create") {
		buttons = append(buttons, []keyboards.Button{{
			Text:    "📋 Клонировать",
			Payload: fmt.Sprintf("role:clone:%s", role.ID.String()),
		}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func formatPermissionMatrix(access domain.JSONAccessMap) string {
	if len(access) == 0 {
		return "Нет разрешений\n"
	}

	opDisplay := opDisplayNames

	var lines []string
	for resourceKey, perms := range access {
		moduleKey := resourceKey
		if len(resourceKey) > 9 && resourceKey[:9] == "resource." {
			moduleKey = resourceKey[9:]
		}

		displayName := moduleKey
		if m, ok := domain.ResourceRegistry[moduleKey]; ok {
			displayName = m.DisplayName
		}

		if wildcard, ok := perms["*"]; ok && wildcard == true {
			lines = append(lines, fmt.Sprintf("  ✅ %s — полный доступ", displayName))
		} else {
			var ops []string
			for op, val := range perms {
				if b, ok := val.(bool); ok && b {
					display := op
					if d, ok2 := opDisplay[op]; ok2 {
						display = d
					}
					ops = append(ops, display)
				}
			}
			if len(ops) > 0 {
				lines = append(lines, fmt.Sprintf("  📋 %s — %s", displayName, strings.Join(ops, ", ")))
			}
		}
	}

	if len(lines) == 0 {
		return "Нет разрешений\n"
	}
	return strings.Join(lines, "\n") + "\n"
}

func handleRoleCreateStart(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "create"); !ok {
		return blocked, nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingRoleData
	userCtx.ContextData = map[string]interface{}{
		"step": "name",
	}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons("Введите название роли (отображаемое имя):", keyboards.GetBackKeyboard()), nil
}

func handleRoleEditStart(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "update"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "role:edit:")
	id, _ := uuid.Parse(roleID)

	role, err := svc.GetRoleByID(id)
	if err != nil {
		return types.ResponseWithButtons("🔑 Роль не найдена.", keyboards.GetBackKeyboard()), nil
	}
	if role.IsSystem {
		return types.ResponseWithButtons("❌ Системную роль нельзя редактировать.", keyboards.GetBackKeyboard()), nil
	}

	ctxData := map[string]interface{}{
		"step":              "name",
		"edit_role_id":      id.String(),
		"edit_display_name": role.DisplayName,
		"role_priority":     float64(role.Priority),
		"role_target_type":  role.TargetType,
	}

	for resourceKey, perms := range role.JSONAccess {
		moduleKey := resourceKey
		if len(resourceKey) > 9 && resourceKey[:9] == "resource." {
			moduleKey = resourceKey[9:]
		}

		var ops []interface{}
		if wildcard, ok := perms["*"]; ok && wildcard == true {
			if m, ok := domain.ResourceRegistry[moduleKey]; ok {
				for _, o := range m.Operations {
					ops = append(ops, o)
				}
			}
		} else {
			for op, val := range perms {
				if b, ok := val.(bool); ok && b {
					ops = append(ops, op)
				}
			}
		}
		if len(ops) > 0 {
			ctxData["perms_"+moduleKey] = ops
			ctxData["mod_"+moduleKey] = true
		}
	}

	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingRoleData
	userCtx.ContextData = ctxData
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons(fmt.Sprintf("Текущее название: %s\n\nВведите новое название роли:", role.DisplayName), keyboards.GetBackKeyboard()), nil
}

func handleRoleCloneStart(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "create"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "role:clone:")
	id, _ := uuid.Parse(roleID)

	role, err := svc.GetRoleByID(id)
	if err != nil {
		return types.ResponseWithButtons("🔑 Роль не найдена.", keyboards.GetBackKeyboard()), nil
	}

	ctxData := map[string]interface{}{
		"step":              "name",
		"clone_from":        role.DisplayName,
		"role_priority":     float64(role.Priority),
		"role_target_type":  role.TargetType,
	}

	for resourceKey, perms := range role.JSONAccess {
		moduleKey := resourceKey
		if len(resourceKey) > 9 && resourceKey[:9] == "resource." {
			moduleKey = resourceKey[9:]
		}

		var ops []interface{}
		if wildcard, ok := perms["*"]; ok && wildcard == true {
			if m, ok := domain.ResourceRegistry[moduleKey]; ok {
				for _, o := range m.Operations {
					ops = append(ops, o)
				}
			}
		} else {
			for op, val := range perms {
				if b, ok := val.(bool); ok && b {
					ops = append(ops, op)
				}
			}
		}
		if len(ops) > 0 {
			ctxData["perms_"+moduleKey] = ops
			ctxData["mod_"+moduleKey] = true
		}
	}

	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingRoleData
	userCtx.ContextData = ctxData
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons(fmt.Sprintf("Клонирование роли «%s».\n\nВведите название новой роли:", role.DisplayName), keyboards.GetBackKeyboard()), nil
}

func HandleRoleMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingRoleData {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)

	switch step {
	case "name":
		userCtx.ContextData["role_name"] = msg
		userCtx.ContextData["step"] = "priority"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите приоритет роли (число):\n\n💡 Приоритет определяет, какая роль главнее. Чем выше число — тем выше приоритет. Если пользователь имеет несколько ролей, действует роль с наибольшим приоритетом.\n\nНапример: Директор — 100, Заведующий общежитием — 50, Староста — 10.", keyboards.GetBackKeyboard()), nil

	case "priority":
		var priority int
		fmt.Sscanf(msg, "%d", &priority)
		userCtx.ContextData["role_priority"] = float64(priority)
		userCtx.ContextData["step"] = "target_type"
		targetButtons := [][]keyboards.Button{
			{
				{Text: "👔 Сотрудник", Payload: "role:target:employee"},
				{Text: "🏠 Житель", Payload: "role:target:resident"},
			},
			{
				{Text: "👥 Оба", Payload: "role:target:both"},
			},
		}
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Выберите, для кого предназначена роль:", targetButtons), nil

	default:
		return nil, nil
	}
}

func HandleRoleCallback(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if keyboards.HasPageSuffix(payload) {
		base := keyboards.StripPageSuffix(payload)
		page := keyboards.ParsePage(payload)
		if strings.HasPrefix(base, "role:mgmt") {
			return handleRoleManagement(ctx, userCtx, svc, fsmMgr, page)
		}
	}

	switch {
	case strings.HasPrefix(payload, "role:view:"):
		return handleRoleView(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:edit:"):
		return handleRoleEditStart(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:clone:"):
		return handleRoleCloneStart(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:delete:"):
		return handleRoleDelete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:confirm_delete:"):
		return handleRoleConfirmDelete(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "role:create":
		return handleRoleCreateStart(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:target:"):
		return handleRoleTargetType(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:module_page:"):
		return handleRoleModulePage(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:select_module:"):
		return handleRoleSelectModule(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:toggle_perm:"):
		return handleRoleTogglePermission(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:save_perms:"):
		return handleRoleSavePerms(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "role:add_more":
		return handleRoleAddMore(ctx, userCtx, svc, fsmMgr)
	case payload == "role:finish":
		return handleRoleFinish(ctx, userCtx, svc, fsmMgr)
	}

	return nil, nil
}

func handleRoleTargetType(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	targetType := strings.TrimPrefix(payload, "role:target:")
	if targetType != domain.TargetTypeEmployee && targetType != domain.TargetTypeResident && targetType != domain.TargetTypeBoth {
		targetType = domain.TargetTypeEmployee
	}

	userCtx.ContextData["role_target_type"] = targetType
	userCtx.ContextData["step"] = "select_module"
	userCtx.ContextData["module_page"] = float64(0)
	modules := buildModuleButtonList(0)
	fsmMgr.Set(ctx, userCtx)
	return types.ResponseWithButtons("Выберите модуль для настройки прав:", modules), nil
}

func handleRoleDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "delete"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "role:delete:")
	return types.ResponseWithButtons("❗️ Точно удалить эту роль?",
		[][]keyboards.Button{
			{{Text: "✅ Да, удалить", Payload: fmt.Sprintf("role:confirm_delete:%s", roleID)}},
			{{Text: "❌ Отмена", Payload: keyboards.PayloadNavBack}},
		}), nil
}

func handleRoleConfirmDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "role", "delete"); !ok {
		return blocked, nil
	}

	roleID := strings.TrimPrefix(payload, "role:confirm_delete:")
	id, _ := uuid.Parse(roleID)

	if err := svc.DeleteRole(id); err != nil {
		log.Printf("Error deleting role: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	userCtx.State = fsm.StateRoleManagement
	fsmMgr.Set(ctx, userCtx)

	return types.ResponseWithButtons("✅ Роль удалена.", keyboards.GetBackKeyboard()), nil
}

func handleRoleModulePage(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	pageStr := strings.TrimPrefix(payload, "role:module_page:")
	page := 0
	fmt.Sscanf(pageStr, "%d", &page)

	userCtx.ContextData["module_page"] = float64(page)
	userCtx.ContextData["step"] = "select_module"
	modules := buildModuleButtonList(page)
	fsmMgr.Set(ctx, userCtx)
	return types.ResponseWithButtons("Выберите модуль для настройки прав:", modules), nil
}

func handleRoleSelectModule(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	moduleKey := strings.TrimPrefix(payload, "role:select_module:")

	userCtx.ContextData["step"] = "select_permissions"
	userCtx.ContextData["current_module"] = moduleKey
	fsmMgr.Set(ctx, userCtx)

	rm, ok := domain.ResourceRegistry[moduleKey]
	if !ok {
		return types.ResponseWithButtons("❌ Неизвестный модуль.", keyboards.GetBackKeyboard()), nil
	}

	existingPerms := getPermsFromCtx(userCtx.ContextData, moduleKey)
	permSet := make(map[string]bool)
	for _, p := range existingPerms {
		permSet[p] = true
	}

	buttons := [][]keyboards.Button{}
	for _, op := range rm.Operations {
		checkmark := "⬜"
		if permSet[op] {
			checkmark = "✅"
		}
		opDisplay := opDisplayNames
		display := op
		if d, ok := opDisplay[op]; ok {
			display = d
		}
		buttons = append(buttons, []keyboards.Button{{
			Text:    fmt.Sprintf("%s %s", checkmark, display),
			Payload: fmt.Sprintf("role:toggle_perm:%s:%s", moduleKey, op),
		}})
	}

	haveAll := true
	for _, op := range rm.Operations {
		if !permSet[op] {
			haveAll = false
			break
		}
	}
	checkmark := "⬜"
	if haveAll {
		checkmark = "✅"
	}
	buttons = append(buttons, []keyboards.Button{{
		Text:    fmt.Sprintf("%s Всё", checkmark),
		Payload: fmt.Sprintf("role:toggle_perm:%s:*", moduleKey),
	}})

	buttons = append(buttons, []keyboards.Button{
		{Text: "✅ Готово", Payload: fmt.Sprintf("role:save_perms:%s", moduleKey)},
	})

	return types.ResponseWithButtons(fmt.Sprintf("Модуль: %s\n\n💡 Права доступа:\n• Просмотр — только смотреть\n• Добавить — создавать записи\n• Изменить — редактировать\n• Удалить — удалять записи\n• Всё — полный доступ\n\nВыберите нужные права:", rm.DisplayName), buttons), nil
}

func handleRoleTogglePermission(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	payload = strings.TrimPrefix(payload, "role:toggle_perm:")
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		return nil, nil
	}
	moduleKey, op := parts[0], parts[1]

	existing := getPermsFromCtx(userCtx.ContextData, moduleKey)

	if op == "*" {
		rm, ok := domain.ResourceRegistry[moduleKey]
		if ok {
			haveAll := len(existing) == len(rm.Operations)
			if haveAll {
				setPermsInCtx(userCtx.ContextData, moduleKey, nil)
			} else {
				setPermsInCtx(userCtx.ContextData, moduleKey, rm.Operations)
			}
		}
	} else {
		found := false
		var newPerms []string
		for _, p := range existing {
			if p == op {
				found = true
			} else {
				newPerms = append(newPerms, p)
			}
		}
		if !found {
			newPerms = append(newPerms, op)
		}
		setPermsInCtx(userCtx.ContextData, moduleKey, newPerms)
	}

	fsmMgr.Set(ctx, userCtx)

	return handleRoleSelectModule(ctx, userCtx, fmt.Sprintf("role:select_module:%s", moduleKey), svc, fsmMgr)
}

func handleRoleSavePerms(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	moduleKey := strings.TrimPrefix(payload, "role:save_perms:")
	perms := getPermsFromCtx(userCtx.ContextData, moduleKey)

	if len(perms) > 0 {
		userCtx.ContextData["mod_"+moduleKey] = true
	} else {
		delete(userCtx.ContextData, "mod_"+moduleKey)
		delete(userCtx.ContextData, "perms_"+moduleKey)
	}

	userCtx.ContextData["step"] = "add_more"

	builtPerms := buildCurrentPermsFromCtx(userCtx)
	preview := formatPermissionMatrix(builtPerms)

	fsmMgr.Set(ctx, userCtx)
	return types.ResponseWithButtons(
		fmt.Sprintf("Права для модуля сохранены.\n\nТекущий набор прав:\n%s\n\nДобавить ещё модуль?", preview),
		[][]keyboards.Button{
			{{Text: "➕ Да, добавить модуль", Payload: "role:add_more"}, {Text: "✅ Завершить", Payload: "role:finish"}},
		},
	), nil
}

func handleRoleAddMore(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.ContextData["step"] = "select_module"
	userCtx.ContextData["module_page"] = float64(0)
	modules := buildModuleButtonList(0)
	fsmMgr.Set(ctx, userCtx)
	return types.ResponseWithButtons("Выберите следующий модуль:", modules), nil
}

func handleRoleFinish(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	roleName, _ := userCtx.ContextData["role_name"].(string)
	if roleName == "" {
		return types.ResponseWithButtons("❌ Название роли не задано.", keyboards.GetBackKeyboard()), nil
	}

	targetType, _ := userCtx.ContextData["role_target_type"].(string)
	if targetType == "" {
		targetType = domain.TargetTypeEmployee
	}

	rolePriority := 0
	if p, ok := userCtx.ContextData["role_priority"].(float64); ok {
		rolePriority = int(p)
	}

	builtAccess := buildCurrentPermsFromCtx(userCtx)

	if errors := domain.ValidateJSONAccess(builtAccess); len(errors) > 0 {
		return types.ResponseWithButtons("❌ Ошибки в правах:\n• "+strings.Join(errors, "\n• "), keyboards.GetBackKeyboard()), nil
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	eid, _ := uuid.Parse(userCtx.EmployeeUUID)
	allowedPerms, err := svc.GetEffectivePermissions(did, eid)
	if err == nil {
		for resourceKey, requestedPerms := range builtAccess {
			allowedForResource, hasAccess := allowedPerms[resourceKey]
			displayName := strings.TrimPrefix(resourceKey, "resource.")
			if !hasAccess {
				return types.ResponseWithButtons(fmt.Sprintf("❌ У вас нет доступа к модулю «%s».\nНельзя выдать права, которых у вас нет.", displayName), keyboards.GetBackKeyboard()), nil
			}
			if _, hasWildcard := allowedForResource["*"]; hasWildcard {
				continue
			}
			if _, reqWildcard := requestedPerms["*"]; reqWildcard {
				return types.ResponseWithButtons(fmt.Sprintf("❌ Нельзя выдать полный доступ к «%s».\nУ вас нет полного доступа к этому модулю.", displayName), keyboards.GetBackKeyboard()), nil
			}
		}
	}

	editRoleID, isEdit := userCtx.ContextData["edit_role_id"].(string)

	if isEdit {
		id, _ := uuid.Parse(editRoleID)
		if _, err := svc.UpdateRole(id, roleName, targetType, rolePriority, builtAccess); err != nil {
			log.Printf("Error updating role: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
	} else {
		if _, err := svc.CreateRole(roleName, roleName, targetType, rolePriority, builtAccess, false); err != nil {
			log.Printf("Error creating role: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
	}

	userCtx.ContextData = nil
	userCtx.ClearNavigationStack()
	userCtx.State = fsm.StateRoleVerified
	userCtx.PushState()
	userCtx.State = fsm.StateRoleManagement
	fsmMgr.Set(ctx, userCtx)

	action := "создана"
	if isEdit {
		action = "обновлена"
	}
	return types.ResponseWithButtons(fmt.Sprintf("✅ Роль «%s» %s.", roleName, action), keyboards.GetBackKeyboard()), nil
}

func buildCurrentPermsFromCtx(userCtx *fsm.UserContext) domain.JSONAccessMap {
	access := make(domain.JSONAccessMap)

	for key, val := range userCtx.ContextData {
		if !strings.HasPrefix(key, "mod_") {
			continue
		}
		if b, ok := val.(bool); !ok || !b {
			continue
		}

		moduleKey := strings.TrimPrefix(key, "mod_")
		perms := getPermsFromCtx(userCtx.ContextData, moduleKey)
		if len(perms) == 0 {
			continue
		}

		permMap := make(domain.ResourcePermissions)
		rm, ok := domain.ResourceRegistry[moduleKey]
		if ok && len(perms) == len(rm.Operations) {
			permMap["*"] = true
		} else {
			for _, p := range perms {
				permMap[p] = true
			}
		}
		access["resource."+moduleKey] = permMap
	}
	return access
}

func getPermsFromCtx(ctxData map[string]interface{}, moduleKey string) []string {
	if v, ok := ctxData["perms_"+moduleKey].([]interface{}); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func setPermsInCtx(ctxData map[string]interface{}, moduleKey string, perms []string) {
	if len(perms) == 0 {
		delete(ctxData, "perms_"+moduleKey)
		return
	}
	permList := make([]interface{}, len(perms))
	for i, p := range perms {
		permList[i] = p
	}
	ctxData["perms_"+moduleKey] = permList
}

func buildModuleButtonList(page int) [][]keyboards.Button {
	var modules []domain.ResourceModule
	for _, key := range domain.ResourceRegistryOrder {
		if m, ok := domain.ResourceRegistry[key]; ok {
			modules = append(modules, m)
		}
	}

	pageSize := 6
	start := page * pageSize
	end := start + pageSize
	if end > len(modules) {
		end = len(modules)
	}
	if start >= len(modules) {
		return [][]keyboards.Button{{{Text: "⬅️ Назад", Payload: "role:add_more"}}}
	}

	var buttons [][]keyboards.Button
	for i := start; i < end; i++ {
		m := modules[i]
		buttons = append(buttons, []keyboards.Button{{
			Text:    m.DisplayName,
			Payload: "role:select_module:" + m.Key,
		}})
	}

	var navRow []keyboards.Button
	if page > 0 {
		navRow = append(navRow, keyboards.Button{Text: "⬅️ Назад", Payload: fmt.Sprintf("role:module_page:%d", page-1)})
	}
	if end < len(modules) {
		navRow = append(navRow, keyboards.Button{Text: "Далее ▶", Payload: fmt.Sprintf("role:module_page:%d", page+1)})
	}
	if len(navRow) > 0 {
		buttons = append(buttons, navRow)
	}

	buttons = append(buttons, []keyboards.Button{{Text: "✅ Завершить", Payload: "role:finish"}})
	return buttons
}
