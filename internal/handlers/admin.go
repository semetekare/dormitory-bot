package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/keyboards"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/templates"
	"github.com/dormitory-bot/internal/types"
	"github.com/google/uuid"
)

func HandleAdminCallback(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	log.Printf("Admin %d: payload=%s (state=%s)", userCtx.UserID, payload, userCtx.State)

	if keyboards.HasPageSuffix(payload) {
		base := keyboards.StripPageSuffix(payload)
		page := keyboards.ParsePage(payload)
		switch {
		case strings.HasPrefix(base, "admin:chat_links"):
			return handleAdminChatLinks(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "admin:reference"):
			return handleAdminReference(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "cleaning:view"):
			return handleCleaningViewAdmin(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "staff:"):
			return HandleStaffCallback(ctx, userCtx, payload, svc, fsmMgr)
		case strings.HasPrefix(base, "role:"):
			return HandleRoleCallback(ctx, userCtx, payload, svc, fsmMgr)
		default:
			return HandleAdminRoomCallback(ctx, userCtx, payload, svc, fsmMgr)
		}
	}

	switch {
	case payload == "menu:modules":
		return handleAdminModules(ctx, userCtx, svc, fsmMgr)
	case payload == "menu:laundry_mgmt":
		return handleAdminLaundry(ctx, userCtx, svc, fsmMgr)
	case payload == "menu:room_mgmt":
		return handleAdminRooms(ctx, userCtx, svc, fsmMgr)
	case payload == "menu:cleaning_mgmt":
		return handleAdminCleaning(ctx, userCtx, svc, fsmMgr)
	case payload == "menu:chat_links_mgmt":
		return handleAdminChatLinks(ctx, userCtx, svc, fsmMgr, 0)
	case payload == "menu:reference_mgmt":
		return handleAdminReference(ctx, userCtx, svc, fsmMgr, 0)
	case payload == "menu:staff_mgmt":
		return handleStaffManagement(ctx, userCtx, svc, fsmMgr, 0)
	case payload == "menu:role_mgmt":
		return handleRoleManagement(ctx, userCtx, svc, fsmMgr, 0)
	case payload == "reference:create":
		return handleReferenceCreateStart(ctx, userCtx, svc, fsmMgr)
	case payload == "chat_link:create":
		return handleChatLinkCreateStart(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "chat_link:view:"):
		return handleChatLinkView(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "chat_link:delete:"):
		return handleChatLinkDelete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "chat_link:confirm_delete:"):
		return handleChatLinkConfirmDelete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "reference:view:"):
		return handleReferenceView(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "reference:delete:"):
		return handleReferenceDelete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "reference:confirm_delete:"):
		return handleReferenceConfirmDelete(ctx, userCtx, payload, svc, fsmMgr)
	case payload == keyboards.PayloadNavBack:
		return handleAdminNavBack(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "dormitory:"):
		return HandleCallbackQuery(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "module:"):
		return handleReferenceCreateModule(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "platform:"):
		return handlePlatformSelect(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin"):
		return HandleAdminLaundryCallback(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "sanitary_remove:"):
		return handleControlRemove(ctx, userCtx, 0, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "discipline_remove:"):
		return handleControlRemove(ctx, userCtx, 1, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cohabitation_remove:"):
		return handleControlRemove(ctx, userCtx, 2, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "room:"),
		strings.HasPrefix(payload, "control:"),
		strings.HasPrefix(payload, "cleaning:"),
		strings.HasPrefix(payload, "penalty:"),
		strings.HasPrefix(payload, "exemption:"),
		strings.HasPrefix(payload, "duty:"):
		return HandleAdminRoomCallback(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "debt:"):
		return HandleAdminRoomCallback(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "staff:"):
		return HandleStaffCallback(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "role:"):
		return HandleRoleCallback(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "resident_role:exit":
		return handleResidentRoleExit(ctx, userCtx, svc, fsmMgr)
	default:
		log.Printf("Admin: unhandled payload, falling back to resident handler: %s", payload)
		return HandleCallbackQuery(ctx, userCtx, payload, svc, fsmMgr)
	}
}

func handleAdminNavBack(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	prevState := userCtx.PopState()

	if prevState == fsm.StateAuthorized && userCtx.EmployeeUUID != "" && userCtx.EmployeeRole != "" {
		userCtx.State = fsm.StateRoleVerified
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
		kb := newAdminMenuBuilderFromContext(svc, userCtx)
		return types.ResponseWithButtons(
			templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
			kb.build(ctx),
		), nil
	}

	userCtx.State = prevState
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	switch prevState {
	case fsm.StateAuthorizedEmployee, fsm.StateRoleVerified:
		kb := newAdminMenuBuilderFromContext(svc, userCtx)
		return types.ResponseWithButtons(
			templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
			kb.build(ctx),
		), nil
	case fsm.StateResidentRoleVerified:
		rolenames := buildRoleDisplayLookup(svc)
		return types.ResponseWithButtons(
			fmt.Sprintf("🔑 Режим: %s", service.MapRoleDisplayWithLookup(userCtx.EmployeeRole, rolenames)),
			newAdminMenuBuilderFromContext(svc, userCtx).build(ctx),
		), nil
	case fsm.StateResidentProfile:
		if roomID, ok := userCtx.ContextData["room_id"].(string); ok {
			rid, _ := uuid.Parse(roomID)
			return viewRoom(ctx, userCtx, rid, svc, fsmMgr, 0)
		}
		return handleAdminRooms(ctx, userCtx, svc, fsmMgr)
	case fsm.StateAuthorized:
		if userCtx.EmployeeUUID != "" {
			userCtx.State = fsm.StateRoleVerified
			fsmMgr.Set(ctx, userCtx)
			kb := newAdminMenuBuilderFromContext(svc, userCtx)
			return types.ResponseWithButtons(
				templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
				kb.build(ctx),
			), nil
		}
		return types.ResponseWithButtons(templates.MsgMainMenu, buildResidentMainMenu(ctx, userCtx, svc)), nil
	default:
		if userCtx.EmployeeUUID != "" {
			userCtx.State = fsm.StateRoleVerified
			fsmMgr.Set(ctx, userCtx)
			kb := newAdminMenuBuilderFromContext(svc, userCtx)
			return types.ResponseWithButtons(
				templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
				kb.build(ctx),
			), nil
		}
		if userCtx.EmployeeRole != "" {
			rolenames := buildRoleDisplayLookup(svc)
			return types.ResponseWithButtons(
				fmt.Sprintf("🔑 Режим: %s", service.MapRoleDisplayWithLookup(userCtx.EmployeeRole, rolenames)),
				newAdminMenuBuilderFromContext(svc, userCtx).build(ctx),
			), nil
		}
		return types.ResponseWithButtons(templates.MsgMainMenu, buildResidentMainMenu(ctx, userCtx, svc)), nil
	}
}

func handleAdminModules(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateModulesManagement
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	modules, err := svc.BotModuleService.GetByDormitory(did)
	if err != nil || len(modules) == 0 {
		return types.ResponseWithButtons("🤖 Нет настроенных ботов.\n\nДобавление модулей — в панели управления.", keyboards.GetBackKeyboard()), nil
	}

	var text string
	for _, m := range modules {
		text += fmt.Sprintf("• %s (%s) — %s\n", m.BotName, m.Platform, m.Status)
	}

	return types.ResponseWithButtons(text, keyboards.GetBackKeyboard()), nil
}

func handleAdminRooms(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateRoomManagement
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	floors, err := svc.ResidentService.GetFloorsByDormitory(did)
	if err != nil || len(floors) == 0 {
		return types.ResponseWithButtons("🚪 Нет этажей.", keyboards.GetBackKeyboard()), nil
	}

	var buttons [][]keyboards.Button
	for _, f := range floors {
		rooms, _ := svc.ResidentService.GetRoomsByFloor(f.ID)
		buttons = append(buttons, []keyboards.Button{{
			Text:    fmt.Sprintf("🏢 Этаж %d (%d комнат)", f.FloorNumber, len(rooms)),
			Payload: fmt.Sprintf("room:floor:%s", f.ID.String()),
		}})
	}

	msg := fmt.Sprintf("🚪 Общежитие — %d этажей\n\nВыбери этаж:", len(floors))
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleAdminCleaning(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	return showAdminCleaning(ctx, userCtx, svc, fsmMgr, nil)
}

func showAdminCleaning(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, monthOffset *int) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateCleaningManagement
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	now := time.Now()
	m := int(now.Month())
	y := now.Year()

	if monthOffset != nil {
		m += *monthOffset
		for m > 12 {
			m -= 12
			y++
		}
		for m < 1 {
			m += 12
			y--
		}
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	duties, err := svc.CleaningService.GetDormitorySchedule(did, m, y)

	monthNames := []string{"", "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"}

	buttons := [][]keyboards.Button{}

	navRow := []keyboards.Button{
		{Text: "◀", Payload: "cleaning:month:-1"},
		{Text: fmt.Sprintf("%s %d", monthNames[m], y), Payload: "cleaning:month:0"},
		{Text: "▶", Payload: "cleaning:month:1"},
	}

	if err != nil || len(duties) == 0 {
		if hasPermission(userCtx, svc, "cleaning", "create") {
			buttons = append(buttons, []keyboards.Button{{Text: "🔄 Сгенерировать график", Payload: "cleaning:generate"}})
		}
		buttons = append(buttons, navRow)
		buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
		return types.ResponseWithButtons(
			fmt.Sprintf("🧹 Дежурства — %s %d\n\nГрафик не сгенерирован.", monthNames[m], y),
			buttons,
		), nil
	}

	statusCounts := map[string]int{}
	for _, d := range duties {
		statusCounts[d.Status]++
	}
	msg := fmt.Sprintf("🧹 Дежурства — %s %d\n\nВсего: %d | ✅ %d | ❌ %d | ⏳ %d",
		monthNames[m], y, len(duties), statusCounts["completed"], statusCounts["missed"], statusCounts["scheduled"])

	buttons = append(buttons, []keyboards.Button{{Text: fmt.Sprintf("📋 Просмотреть (%d)", len(duties)), Payload: "cleaning:view:0"}})
	buttons = append(buttons, []keyboards.Button{{Text: "🔄 Перегенерировать", Payload: "cleaning:generate"}})
	buttons = append(buttons, navRow)
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleCleaningMonth(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	offsetStr := strings.TrimPrefix(payload, "cleaning:month:")
	offset, _ := strconv.Atoi(offsetStr)
	return showAdminCleaning(ctx, userCtx, svc, fsmMgr, &offset)
}

func handleAdminChatLinks(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if page == 0 {
		userCtx.PushState()
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	links, err := svc.ChatLinkService.List(did)

	if err != nil || len(links) == 0 {
		return types.ResponseWithButtons("💬 Ссылок на чаты пока нет.\n\nНажмите «Создать» для добавления.", [][]keyboards.Button{{{Text: "➕ Создать", Payload: "chat_link:create"}, {Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}}}), nil
	}

	items := make([]keyboards.PageableButton, 0, len(links))
	for _, l := range links {
		items = append(items, keyboards.PageableButton{
			Text:    fmt.Sprintf("%s (%s)", truncate(l.Title, 25), truncate(l.URL, 15)),
			Payload: fmt.Sprintf("chat_link:view:%s", l.ID.String()),
		})
	}

	msg := "💬 Ссылки на чаты (Управление):\n\n"
	for _, l := range links {
		msg += fmt.Sprintf("• %s\n  %s (%s)\n", l.Title, l.URL, l.Platform)
	}
	msg += fmt.Sprintf("\nВсего: %d", len(links))

	kb := keyboards.BuildPaginatedKeyboard(items, "admin:chat_links", page, keyboards.DefaultPageSize)
	kb[len(kb)-1] = []keyboards.Button{{Text: "➕ Создать", Payload: "chat_link:create"}, {Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}}

	return types.ResponseWithButtons(msg, kb), nil
}

func handleChatLinkView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	linkID := strings.TrimPrefix(payload, "chat_link:view:")
	id, _ := uuid.Parse(linkID)
	link, err := svc.ChatLinkService.GetByID(id)
	if err != nil {
		return types.ResponseWithButtons("💬 Ссылка не найдена.", keyboards.GetBackKeyboard()), nil
	}
	msg := fmt.Sprintf("💬 %s\n\n📎 %s\n🖥 %s", link.Title, link.URL, link.Platform)
	return types.ResponseWithButtons(msg, [][]keyboards.Button{
		{{Text: "🗑️ Удалить", Payload: fmt.Sprintf("chat_link:delete:%s", link.ID.String())}},
		{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
	}), nil
}

func handleChatLinkDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "chat_link", "delete"); !ok {
		return blocked, nil
	}
	linkID := strings.TrimPrefix(payload, "chat_link:delete:")
	return types.ResponseWithButtons("❗️ Точно удалить эту ссылку?",
		[][]keyboards.Button{
			{{Text: "✅ Да, удалить", Payload: fmt.Sprintf("chat_link:confirm_delete:%s", linkID)}},
			{{Text: "❌ Отмена", Payload: keyboards.PayloadNavBack}},
		}), nil
}

func handleChatLinkConfirmDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	linkID := strings.TrimPrefix(payload, "chat_link:confirm_delete:")
	id, _ := uuid.Parse(linkID)
	if err := svc.ChatLinkService.Delete(id); err != nil {
		log.Printf("Error deleting chat link: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return handleAdminChatLinks(ctx, userCtx, svc, fsmMgr, 0)
}

func handleAdminReference(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if page == 0 {
		userCtx.PushState()
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	materials, err := svc.ReferenceService.GetByDormitory(did)

	if err != nil || len(materials) == 0 {
		return types.ResponseWithButtons("📖 Справочных материалов пока нет.\n\nНажмите «Создать» для добавления.", [][]keyboards.Button{{{Text: "➕ Создать", Payload: "reference:create"}, {Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}}}), nil
	}

	items := make([]keyboards.PageableButton, 0, len(materials))
	for _, m := range materials {
		module := ""
		if m.ModuleKey != nil {
			module = *m.ModuleKey
		}
		items = append(items, keyboards.PageableButton{
			Text:    fmt.Sprintf("%s [%s]", truncate(m.Name, 25), module),
			Payload: fmt.Sprintf("reference:view:%s", m.ID.String()),
		})
	}

	msg := "📖 Справочные материалы (Управление):\n\n"
	for _, m := range materials {
		module := ""
		if m.ModuleKey != nil {
			module = *m.ModuleKey
		}
		msg += fmt.Sprintf("• %s [%s] (Модуль: %s)\n", m.Name, m.Category, module)
	}
	msg += fmt.Sprintf("\nВсего: %d", len(materials))

	kb := keyboards.BuildPaginatedKeyboard(items, "admin:reference", page, keyboards.DefaultPageSize)
	kb[len(kb)-1] = []keyboards.Button{{Text: "➕ Создать", Payload: "reference:create"}, {Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}}

	return types.ResponseWithButtons(msg, kb), nil
}

func handleReferenceView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	refID := strings.TrimPrefix(payload, "reference:view:")
	id, _ := uuid.Parse(refID)
	material, err := svc.ReferenceService.GetByID(id)
	if err != nil {
		return types.ResponseWithButtons("📖 Справка не найдена.", keyboards.GetBackKeyboard()), nil
	}
	module := ""
	if material.ModuleKey != nil {
		module = *material.ModuleKey
	}
	msg := fmt.Sprintf("📖 %s\n\nКатегория: %s\nМодуль: %s\n\n%s", material.Name, material.Category, module, material.Description)
	return types.ResponseWithButtons(msg, [][]keyboards.Button{
		{{Text: "🗑️ Удалить", Payload: fmt.Sprintf("reference:delete:%s", material.ID.String())}},
		{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
	}), nil
}

func handleReferenceDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "reference", "delete"); !ok {
		return blocked, nil
	}
	refID := strings.TrimPrefix(payload, "reference:delete:")
	return types.ResponseWithButtons("❗️ Точно удалить эту справку?",
		[][]keyboards.Button{
			{{Text: "✅ Да, удалить", Payload: fmt.Sprintf("reference:confirm_delete:%s", refID)}},
			{{Text: "❌ Отмена", Payload: keyboards.PayloadNavBack}},
		}), nil
}

func handleReferenceConfirmDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	refID := strings.TrimPrefix(payload, "reference:confirm_delete:")
	id, _ := uuid.Parse(refID)
	if err := svc.ReferenceService.Delete(id); err != nil {
		log.Printf("Error deleting reference: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return handleAdminReference(ctx, userCtx, svc, fsmMgr, 0)
}

func handleReferenceCreateStart(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "reference", "create"); !ok {
		return blocked, nil
	}
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingReferenceData
	userCtx.ContextData = map[string]interface{}{"step": "name"}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите название справки:", keyboards.GetBackKeyboard()), nil
}
func HandleAdminMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State == fsm.StateAwaitingReferenceData {
		return handleReferenceCreateProcess(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingChatLinkData {
		return handleChatLinkCreateProcess(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingControlData {
		return HandleAdminControlMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingMachineData {
		return HandleAdminLaundryMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingLaundrySettings {
		return HandleAdminLaundrySettingsMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingAdminBooking {
		return HandleAdminBookingMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingPenaltyData {
		return HandleAdminPenaltyMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingExemptionData {
		return HandleAdminExemptionMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingStaffData {
		return HandleStaffMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	if userCtx.State == fsm.StateAwaitingRoleData {
		return HandleRoleMessage(ctx, userCtx, msg, svc, fsmMgr)
	}
	return nil, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func handleReferenceCreateProcess(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	step := userCtx.ContextData["step"].(string)

	switch step {
	case "name":
		userCtx.ContextData["name"] = msg
		userCtx.ContextData["step"] = "description"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите описание справки:", keyboards.GetBackKeyboard()), nil
	case "description":
		userCtx.ContextData["description"] = msg
		userCtx.ContextData["step"] = "category"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите категорию:", keyboards.GetBackKeyboard()), nil
	case "category":
		userCtx.ContextData["category"] = msg
		userCtx.ContextData["step"] = "module"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Выберите модуль для справки:", keyboards.GetModuleSelectKeyboard()), nil
	}
	return nil, nil
}

func handleChatLinkCreateStart(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "chat_link", "create"); !ok {
		return blocked, nil
	}
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingChatLinkData
	userCtx.ContextData = map[string]interface{}{"step": "title"}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите название ссылки:", keyboards.GetBackKeyboard()), nil
}

func handleChatLinkCreateProcess(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	step := userCtx.ContextData["step"].(string)

	switch step {
	case "title":
		userCtx.ContextData["title"] = msg
		userCtx.ContextData["step"] = "url"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите URL ссылки:", keyboards.GetBackKeyboard()), nil
	case "url":
		userCtx.ContextData["url"] = msg
		userCtx.ContextData["step"] = "platform"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Выберите платформу:", keyboards.GetPlatformSelectKeyboard()), nil
	}
	return nil, nil
}

func handleReferenceCreateModule(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "reference", "create"); !ok {
		return blocked, nil
	}
	if userCtx.State != fsm.StateAwaitingReferenceData {
		return nil, nil
	}

	module := strings.TrimPrefix(payload, "module:")

	did, _ := uuid.Parse(userCtx.DormitoryID)
	ref := &domain.ReferenceMaterial{
		DormitoryID: did,
		Name:        userCtx.ContextData["name"].(string),
		Description: userCtx.ContextData["description"].(string),
		Category:    userCtx.ContextData["category"].(string),
		CreatedBy:   uuid.MustParse(userCtx.EmployeeUUID),
		ModuleKey:   &module,
	}

	if err := svc.ReferenceService.Create(ref); err != nil {
		log.Printf("Error creating reference: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	userCtx.ContextData = nil
	userCtx.State = fsm.StateRoleVerified
	fsmMgr.Set(ctx, userCtx)
	return handleAdminReference(ctx, userCtx, svc, fsmMgr, 0)
}

func handlePlatformSelect(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingChatLinkData {
		return nil, nil
	}

	platform := strings.TrimPrefix(payload, "platform:")

	var platformName string
	switch platform {
	case "max":
		platformName = "max"
	case "vk":
		platformName = "vk"
	case "other":
		platformName = "other"
	default:
		return types.ResponseWithButtons("❌ Неверная платформа.", keyboards.GetBackKeyboard()), nil
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)

	link := &domain.ChatLink{
		DormitoryID: did,
		Title:       userCtx.ContextData["title"].(string),
		URL:         userCtx.ContextData["url"].(string),
		Platform:    platformName,
		LinkType:    "dormitory",
		CreatedBy:   uuid.MustParse(userCtx.EmployeeUUID),
	}
	if err := svc.ChatLinkService.Create(link); err != nil {
		log.Printf("Error creating chat link: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	userCtx.ContextData = nil
	userCtx.State = fsm.StateRoleVerified
	fsmMgr.Set(ctx, userCtx)
	return handleAdminChatLinks(ctx, userCtx, svc, fsmMgr, 0)
}
