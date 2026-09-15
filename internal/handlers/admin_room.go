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
	"github.com/dormitory-bot/internal/timeutil"
	"github.com/dormitory-bot/internal/types"
	"github.com/google/uuid"
)

func HandleAdminRoomCallback(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	log.Printf("Admin %d: Room callback: %s", userCtx.UserID, payload)

	if keyboards.HasPageSuffix(payload) {
		base := keyboards.StripPageSuffix(payload)
		page := keyboards.ParsePage(payload)
		switch {
		case strings.HasPrefix(base, "room:floor:"):
			return handleRoomFloorView(ctx, userCtx, base, svc, fsmMgr, page)
		case strings.HasPrefix(base, "room:view:"):
			return handleRoomView(ctx, userCtx, base, svc, fsmMgr, page)
		case strings.HasPrefix(base, "room:search:"):
			return handleRoomSearch(ctx, userCtx, base, svc, fsmMgr, page)
		case strings.HasPrefix(base, "cleaning:view"):
			return handleCleaningViewAdmin(ctx, userCtx, svc, fsmMgr, page)
		}
	}

	switch {
	case strings.HasPrefix(payload, "room:floor:"):
		return handleRoomFloorView(ctx, userCtx, payload, svc, fsmMgr, 0)
	case strings.HasPrefix(payload, "room:search:"):
		return handleRoomSearch(ctx, userCtx, payload, svc, fsmMgr, 0)
	case strings.HasPrefix(payload, "room:view:"):
		return handleRoomView(ctx, userCtx, payload, svc, fsmMgr, 0)
	case strings.HasPrefix(payload, "room:resident:"):
		return handleResidentProfile(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "control:sanitary:set:"):
		return handleControlSet(ctx, userCtx, ctrlSanitary, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "control:discipline:set:"):
		return handleControlSet(ctx, userCtx, ctrlDiscipline, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "control:cohabitation:set:"):
		return handleControlSet(ctx, userCtx, ctrlCohabitation, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "sanitary_remove:"):
		return handleControlRemove(ctx, userCtx, ctrlSanitary, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "discipline_remove:"):
		return handleControlRemove(ctx, userCtx, ctrlDiscipline, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cohabitation_remove:"):
		return handleControlRemove(ctx, userCtx, ctrlCohabitation, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "room:controls:"):
		return handleRoomControlsMenu(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "room:residents:"):
		return handleRoomResidentsMenu(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:generate_months:"):
		return handleCleaningGenerateMonths(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:generate"):
		return handleCleaningGenerate(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:view:"):
		return handleCleaningViewAdmin(ctx, userCtx, svc, fsmMgr, 0)
	case strings.HasPrefix(payload, "cleaning:month:"):
		return handleCleaningMonth(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:complete:"):
		return handleCleaningComplete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:miss:"):
		return handleCleaningMiss(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "duty:view:"):
		return handleDutyView(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "penalty:create":
		return handlePenaltyCreate(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "penalty:from_miss:"):
		return handlePenaltyCreateFromMiss(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "penalty:delete:"):
		return handlePenaltyDelete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "penalty:edit:"):
		return handlePenaltyEdit(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "exemption:create":
		return handleExemptionCreate(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:transfer:"):
		return handleCleaningTransferRoom(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "debt:view:"):
		return handleResidentDebt(ctx, userCtx, payload, svc, fsmMgr)
	case payload == keyboards.PayloadNavBack:
		return handleAdminNavBack(ctx, userCtx, svc, fsmMgr)
	default:
		log.Printf("AdminRoom: unhandled payload, delegating to admin handler: %s", payload)
		return HandleAdminCallback(ctx, userCtx, payload, svc, fsmMgr)
	}
}

func handleRoomSearch(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	query := strings.TrimPrefix(payload, "room:search:")
	did, _ := uuid.Parse(userCtx.DormitoryID)

	rooms, err := svc.ResidentService.SearchRooms(did, query)
	if err != nil || len(rooms) == 0 {
		return types.ResponseWithButtons("🚪 Комнат не найдено.", keyboards.GetBackKeyboard()), nil
	}

	items := make([]keyboards.PageableButton, 0, len(rooms))
	for _, r := range rooms {
		items = append(items, keyboards.PageableButton{
			Text:    fmt.Sprintf("🚪 Комната %s (этаж %d)", r.RoomNumber, r.FloorID),
			Payload: fmt.Sprintf("room:view:%s", r.ID.String()),
		})
	}

	return types.ResponseWithButtons("🚪 Результаты поиска:", keyboards.BuildPaginatedKeyboard(items, fmt.Sprintf("room:search:%s", query), page, keyboards.DefaultPageSize)), nil
}

func handleRoomView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	roomID := strings.TrimPrefix(payload, "room:view:")
	rid, _ := uuid.Parse(roomID)
	return viewRoom(ctx, userCtx, rid, svc, fsmMgr, page)
}

func handleResidentProfile(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	residentID := strings.TrimPrefix(payload, "room:resident:")
	rid, _ := uuid.Parse(residentID)

	profile, err := svc.ResidentService.GetFullProfile(rid)
	if err != nil {
		return types.TextResponse("👤 Жилец не найден."), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateResidentProfile
	fsmMgr.Set(ctx, userCtx)

	msg := svc.ResidentService.FormatResidentProfile(profile)

	if profile.PlatformUserID != "" {
		fullName := fmt.Sprintf("%s %s", profile.User.FirstName, profile.User.LastName)
		if mention := keyboards.FormatMAXMention(profile.PlatformUserID, fullName); mention != "" {
			msg += fmt.Sprintf("\n💬 %s", mention)
		} else {
			msg += "\n⚠️ Этот пользователь ещё не заходил в бота"
		}
	}

	var buttons [][]keyboards.Button
	buttons = append(buttons, []keyboards.Button{{Text: "💰 Задолженности", Payload: fmt.Sprintf("debt:view:%s", residentID)}})
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleCleaningGenerate(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "create"); !ok {
		return blocked, nil
	}

	// Show choice of how many months.
	return types.ResponseWithButtons("🧹 Генерация графика дежурств:\n\nВыберите на сколько месяцев сгенерировать:", [][]keyboards.Button{
		{{Text: "📅 Текущий месяц", Payload: "cleaning:generate_months:1"}},
		{{Text: "📅 На 3 месяца", Payload: "cleaning:generate_months:3"}},
		{{Text: "📅 На 6 месяцев", Payload: "cleaning:generate_months:6"}},
		{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
	}), nil
}

func handleCleaningGenerateMonths(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "create"); !ok {
		return blocked, nil
	}
	monthsStr := strings.TrimPrefix(payload, "cleaning:generate_months:")
	months, _ := strconv.Atoi(monthsStr)
	if months < 1 || months > 6 {
		months = 1
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	if err := svc.CleaningService.GenerateScheduleMulti(did, months); err != nil {
		log.Printf("Error generating multi-month schedule: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	return types.ResponseWithButtons(fmt.Sprintf("✅ График дежурств на %d мес. сгенерирован!", months), keyboards.GetBackKeyboard()), nil
}

func handleCleaningViewAdmin(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	did, _ := uuid.Parse(userCtx.DormitoryID)
	now := time.Now()
	month, year := int(now.Month()), now.Year()

	if m, ok := userCtx.ContextData["cleaning_month"].(float64); ok {
		month = int(m)
	}
	if y, ok := userCtx.ContextData["cleaning_year"].(float64); ok {
		year = int(y)
	}

	duties, err := svc.CleaningService.GetDormitorySchedule(did, month, year)
	if err != nil || len(duties) == 0 {
		return types.ResponseWithButtons("🧹 Дежурств на этот месяц нет.", keyboards.GetBackKeyboard()), nil
	}

	floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
	floorNum := make(map[uuid.UUID]int)
	for _, f := range floors {
		floorNum[f.ID] = f.FloorNumber
	}

	roomNums := make(map[uuid.UUID]string)
	for _, f := range floors {
		rooms, _ := svc.ResidentService.GetRoomsByFloor(f.ID)
		for _, r := range rooms {
			roomNums[r.ID] = r.RoomNumber
		}
	}

	items := make([]keyboards.PageableButton, 0, len(duties))
	for _, d := range duties {
		statusEmoji := "⏳"
		switch d.Status {
		case "completed":
			statusEmoji = "✅"
		case "missed":
			statusEmoji = "❌"
		}
		fn := floorNum[d.FloorID]
		rn := roomNums[d.RoomID]
		if rn == "" {
			rn = d.RoomID.String()[:8]
		}
		txt := fmt.Sprintf("%s %s — эт.%d, к.%s", statusEmoji, timeutil.FormatEU(d.DutyDate), fn, rn)
		items = append(items, keyboards.PageableButton{
			Text:    txt,
			Payload: fmt.Sprintf("duty:view:%s", d.ID.String()),
		})
	}

	msg := fmt.Sprintf("🧹 График дежурств на %d.%d (стр. %d):\n", month, year, page+1)
	return types.ResponseWithButtons(msg, keyboards.BuildPaginatedKeyboard(items, "cleaning:view", page, keyboards.DefaultPageSize)), nil
}

func handlePenaltyCreate(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "create"); !ok {
		return blocked, nil
	}
	roomID, _ := userCtx.ContextData["room_id"].(string)
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingPenaltyData
	userCtx.ContextData = map[string]interface{}{"step": "penalty_count", "room_id": roomID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите количество штрафных дежурств (1-3):", keyboards.GetBackKeyboard()), nil
}

func handlePenaltyCreateFromMiss(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "create"); !ok {
		return blocked, nil
	}
	roomID := strings.TrimPrefix(payload, "penalty:from_miss:")
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingPenaltyData
	userCtx.ContextData = map[string]interface{}{"step": "penalty_count", "room_id": roomID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons(
		fmt.Sprintf("⚠️ Назначение штрафа за пропуск дежурства.\nВведите количество штрафных дежурств (1-3):"),
		keyboards.GetBackKeyboard(),
	), nil
}

func HandleAdminPenaltyMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingPenaltyData {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)

	switch step {
	case "penalty_count":
		count, err := strconv.Atoi(msg)
		if err != nil || count < 1 || count > 3 {
			return types.ResponseWithButtons("❌ Введите число от 1 до 3:", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["count"] = float64(count)
		userCtx.ContextData["step"] = "penalty_reason"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите основание штрафа:", keyboards.GetBackKeyboard()), nil

	case "penalty_reason":
		roomID := userCtx.ContextData["room_id"].(string)
		rid, _ := uuid.Parse(roomID)
		count := int(userCtx.ContextData["count"].(float64))
		employeeUUID, _ := uuid.Parse(userCtx.EmployeeUUID)

		_, err := svc.CleaningService.CreatePenalty(rid, nil, count, msg, employeeUUID)
		if err != nil {
			log.Printf("Error creating penalty: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateCleaningManagement
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons(
			fmt.Sprintf("✅ Назначено %d штрафных дежурств.\nОснование: %s", count, msg),
			keyboards.GetBackKeyboard(),
		), nil

	case "penalty_edit_count":
		count, err := strconv.Atoi(msg)
		if err != nil || count < 1 || count > 3 {
			return types.ResponseWithButtons("❌ Введите число от 1 до 3:", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["count"] = float64(count)
		userCtx.ContextData["step"] = "penalty_edit_reason"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите новое основание штрафа:", keyboards.GetBackKeyboard()), nil

	case "penalty_edit_reason":
		penaltyID := userCtx.ContextData["penalty_id"].(string)
		pid, _ := uuid.Parse(penaltyID)
		count := int(userCtx.ContextData["count"].(float64))

		if err := svc.CleaningService.UpdatePenalty(pid, count, msg); err != nil {
			log.Printf("Error updating penalty: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateCleaningManagement
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons(
			fmt.Sprintf("✅ Штраф изменён: %d дежурств.\nОснование: %s", count, msg),
			keyboards.GetBackKeyboard(),
		), nil

	case "transfer_room":
		dutyID := userCtx.ContextData["transfer_duty_id"].(string)
		did, _ := uuid.Parse(dutyID)
		didDorm, _ := uuid.Parse(userCtx.DormitoryID)

		roomNumber := msg
		rooms, _ := svc.ResidentService.SearchRooms(didDorm, roomNumber)
		if len(rooms) == 0 {
			return types.ResponseWithButtons("❌ Комната не найдена. Введите номер ещё раз:", keyboards.GetBackKeyboard()), nil
		}
		newRoom := rooms[0]

		if err := svc.CleaningService.TransferDuty(did, newRoom.ID); err != nil {
			log.Printf("Error transferring duty: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateCleaningManagement
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons(
			fmt.Sprintf("✅ Дежурство перенесено в комнату %s.", newRoom.RoomNumber),
			keyboards.GetBackKeyboard(),
		), nil
	}

	return nil, nil
}

func handlePenaltyDelete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "delete"); !ok {
		return blocked, nil
	}
	penaltyID := strings.TrimPrefix(payload, "penalty:delete:")
	pid, _ := uuid.Parse(penaltyID)
	if err := svc.CleaningService.DeletePenalty(pid); err != nil {
		log.Printf("Error deleting penalty: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return types.ResponseWithButtons("✅ Штраф удалён.", keyboards.GetBackKeyboard()), nil
}

func handlePenaltyEdit(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "update"); !ok {
		return blocked, nil
	}
	penaltyID := strings.TrimPrefix(payload, "penalty:edit:")
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingPenaltyData
	userCtx.ContextData = map[string]interface{}{"step": "penalty_edit_count", "penalty_id": penaltyID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите новое количество штрафных дежурств (1-3):", keyboards.GetBackKeyboard()), nil
}

func handleExemptionCreate(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "create"); !ok {
		return blocked, nil
	}
	roomID, _ := userCtx.ContextData["room_id"].(string)
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingExemptionData
	userCtx.ContextData = map[string]interface{}{"step": "exemption_start", "room_id": roomID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите дату начала освобождения (ДД.ММ.ГГГГ):", keyboards.GetBackKeyboard()), nil
}

func HandleAdminExemptionMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingExemptionData {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)

	switch step {
	case "exemption_start":
		if _, err := time.Parse("02.01.2006", msg); err != nil {
			return types.ResponseWithButtons("❌ Неверный формат. Введите ДД.ММ.ГГГГ:", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["start_date"] = msg
		userCtx.ContextData["step"] = "exemption_end"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите дату окончания (ДД.ММ.ГГГГ):", keyboards.GetBackKeyboard()), nil

	case "exemption_end":
		if _, err := time.Parse("02.01.2006", msg); err != nil {
			return types.ResponseWithButtons("❌ Неверный формат. Введите ДД.ММ.ГГГГ:", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["end_date"] = msg
		userCtx.ContextData["step"] = "exemption_reason"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите причину освобождения:", keyboards.GetBackKeyboard()), nil

	case "exemption_reason":
		roomID := userCtx.ContextData["room_id"].(string)
		rid, _ := uuid.Parse(roomID)
		startDate := userCtx.ContextData["start_date"].(string)
		endDate := userCtx.ContextData["end_date"].(string)
		reason := msg
		employeeUUID, _ := uuid.Parse(userCtx.EmployeeUUID)

		startParsed, _ := time.Parse("02.01.2006", startDate)
		endParsed, _ := time.Parse("02.01.2006", endDate)
		startDB := startParsed.Format("2006-01-02")
		endDB := endParsed.Format("2006-01-02")

		_, err := svc.CleaningService.CreateExemptionSafe(rid, startDB, endDB, reason, employeeUUID)
		if err != nil {
			log.Printf("Error creating exemption: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateCleaningManagement
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons(
			fmt.Sprintf("✅ Комната освобождена от дежурств.\nПериод: %s – %s\nПричина: %s", startDate, endDate, reason),
			keyboards.GetBackKeyboard(),
		), nil
	}

	return nil, nil
}

func handleCleaningComplete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "update"); !ok {
		return blocked, nil
	}
	dutyID := strings.TrimPrefix(payload, "cleaning:complete:")
	did, _ := uuid.Parse(dutyID)
	employeeUUID, _ := uuid.Parse(userCtx.EmployeeUUID)
	if err := svc.CleaningService.MarkDutyCompleted(did, employeeUUID); err != nil {
		log.Printf("Error marking duty completed: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return types.ResponseWithButtons("✅ Отмечено как выполненное.", keyboards.GetBackKeyboard()), nil
}

func handleCleaningMiss(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "update"); !ok {
		return blocked, nil
	}
	dutyID := strings.TrimPrefix(payload, "cleaning:miss:")
	did, _ := uuid.Parse(dutyID)

	if err := svc.CleaningService.MarkDutyMissed(did); err != nil {
		log.Printf("Error marking duty missed: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	var duty domain.CleaningDuty
	svc.CleaningService.DB().First(&duty, "id = ?", did)

	var room domain.Room
	svc.CleaningService.DB().First(&room, "id = ?", duty.RoomID)

	buttons := [][]keyboards.Button{
		{{Text: fmt.Sprintf("⚠️ Назначить штраф к.%s", room.RoomNumber), Payload: fmt.Sprintf("penalty:from_miss:%s", room.ID.String())}},
		{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
	}

	return types.ResponseWithButtons(
		fmt.Sprintf("❌ Дежурство %s (к.%s) отмечено как пропущенное.\n\nНазначить штрафное дежурство?", timeutil.FormatEU(duty.DutyDate), room.RoomNumber),
		buttons,
	), nil
}

func handleDutyView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	dutyID := strings.TrimPrefix(payload, "duty:view:")
	did, _ := uuid.Parse(dutyID)

	var duty domain.CleaningDuty
	if err := svc.CleaningService.DB().First(&duty, "id = ?", did).Error; err != nil {
		return types.ResponseWithButtons("❌ Дежурство не найдено.", keyboards.GetBackKeyboard()), nil
	}

	var room domain.Room
	svc.CleaningService.DB().First(&room, "id = ?", duty.RoomID)

	typeMap := map[string]string{"regular": "🧹 Очередное", "penalty": "⚠️ Штрафное", "kitchen": "🍳 Кухня", "garbage": "🗑️ Мусор"}
	statusMap := map[string]string{"scheduled": "⏳ Запланировано", "completed": "✅ Выполнено", "missed": "❌ Пропущено"}

	msg := fmt.Sprintf("🧹 Дежурство\n\nДата: %s\nТип: %s\nКомната: %s\nСтатус: %s\n",
		timeutil.FormatEU(duty.DutyDate), typeMap[duty.DutyType], room.RoomNumber, statusMap[duty.Status])

	if duty.DutyType == "penalty" && duty.PenaltyCleaningID != nil {
		var penalty domain.PenaltyCleaning
		svc.CleaningService.DB().First(&penalty, "id = ?", *duty.PenaltyCleaningID)
		msg += fmt.Sprintf("\n⚠️ Штраф #%s: %d из %d отбыто\n   Причина: %s",
			penalty.ID.String()[:8], penalty.CompletedCount, penalty.Count, penalty.Reason)
	}

	buttons := [][]keyboards.Button{}
	if duty.Status == "scheduled" {
		if hasPermission(userCtx, svc, "cleaning", "update") {
			buttons = append(buttons, []keyboards.Button{
				{Text: "✅ Выполнено", Payload: fmt.Sprintf("cleaning:complete:%s", duty.ID.String())},
				{Text: "❌ Пропущено", Payload: fmt.Sprintf("cleaning:miss:%s", duty.ID.String())},
			})
		} else {
			buttons = append(buttons, []keyboards.Button{
				{Text: "✅ Выполнено", Payload: fmt.Sprintf("cleaning:complete:%s", duty.ID.String())},
			})
		}
		if hasPermission(userCtx, svc, "cleaning", "update") {
			buttons = append(buttons, []keyboards.Button{
				{Text: "↗️ Перенести в другую комнату", Payload: fmt.Sprintf("cleaning:transfer:%s", duty.ID.String())},
			})
		}
	}
	if duty.DutyType == "penalty" && duty.PenaltyCleaningID != nil && hasPermission(userCtx, svc, "cleaning", "update") {
		buttons = append(buttons, []keyboards.Button{
			{Text: "✏️ Изменить штраф", Payload: fmt.Sprintf("penalty:edit:%s", duty.PenaltyCleaningID.String())},
			{Text: "🗑️ Удалить штраф", Payload: fmt.Sprintf("penalty:delete:%s", duty.PenaltyCleaningID.String())},
		})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleCleaningTransferRoom(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "cleaning", "update"); !ok {
		return blocked, nil
	}
	dutyID := strings.TrimPrefix(payload, "cleaning:transfer:")
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingPenaltyData
	userCtx.ContextData = map[string]interface{}{"step": "transfer_room", "transfer_duty_id": dutyID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("↗️ Перенос дежурства.\n\nВведите номер комнаты, в которую перенести дежурство:", keyboards.GetBackKeyboard()), nil
}

func handleRoomFloorView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	floorID := strings.TrimPrefix(payload, "room:floor:")
	fid, _ := uuid.Parse(floorID)

	rooms, err := svc.ResidentService.GetRoomsByFloor(fid)
	if err != nil || len(rooms) == 0 {
		return types.ResponseWithButtons("🚪 Комнат на этаже нет.", keyboards.GetBackKeyboard()), nil
	}

	items := make([]keyboards.PageableButton, 0, len(rooms))
	for _, r := range rooms {
		residents, _ := svc.RoomService.GetRoomResidents(r.ID)
		label := fmt.Sprintf("🚪 %s (%d/%d чел)", r.RoomNumber, len(residents), r.Capacity)
		if !r.IsActive {
			label += " ⛔"
		}
		items = append(items, keyboards.PageableButton{
			Text:    label,
			Payload: fmt.Sprintf("room:view:%s", r.ID.String()),
		})
	}

	return types.ResponseWithButtons("🚪 Комнаты на этаже:", keyboards.BuildPaginatedKeyboard(items, fmt.Sprintf("room:floor:%s", floorID), page, keyboards.DefaultPageSize)), nil
}

func handleResidentDebt(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	residentID := strings.TrimPrefix(payload, "debt:view:")
	rid, _ := uuid.Parse(residentID)

	debts, err := svc.ResidentService.GetDebt(rid)
	if err != nil || len(debts) == 0 {
		return types.ResponseWithButtons("💰 Задолженностей нет.", keyboards.GetBackKeyboard()), nil
	}

	msg := "💰 Задолженности жильца:\n\n"
	for _, d := range debts {
		msg += fmt.Sprintf("• %s: %.2f руб. — %s\n", d.Period, d.Amount, d.Description)
	}

	return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil
}

func handleRoomControlsMenu(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	roomID := strings.TrimPrefix(payload, "room:controls:")
	return types.ResponseWithButtons("🔍 Выберите тип контроля:",
		[][]keyboards.Button{
			{{Text: "🧹 Санконтроль", Payload: fmt.Sprintf("control:sanitary:set:%s", roomID)}},
			{{Text: "⚠️ Дисципл. контроль", Payload: fmt.Sprintf("control:discipline:set:%s", roomID)}},
			{{Text: "🏠 Контроль проживания", Payload: fmt.Sprintf("control:cohabitation:set:%s", roomID)}},
			{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
		}), nil
}

func handleRoomResidentsMenu(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	roomID := strings.TrimPrefix(payload, "room:residents:")
	rid, _ := uuid.Parse(roomID)
	residents, _ := svc.RoomService.GetRoomResidents(rid)

	if len(residents) == 0 {
		return types.ResponseWithButtons("👤 В этой комнате никто не проживает.", keyboards.GetBackKeyboard()), nil
	}

	buttons := [][]keyboards.Button{}
	for _, r := range residents {
		buttons = append(buttons, []keyboards.Button{{
			Text:    fmt.Sprintf("👤 %s %s", r.User.LastName, r.User.FirstName),
			Payload: fmt.Sprintf("room:resident:%s", r.ID.String()),
		}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons("👤 Жильцы комнаты:", buttons), nil
}
