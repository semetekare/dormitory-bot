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

func HandleAdminLaundryCallback(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	switch {
	case payload == "laundry:admin_machines":
		return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)
	case payload == "laundry:admin_create_machine":
		return handleAdminCreateMachineStart(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_edit_machine:"):
		return handleAdminEditMachineStart(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_delete_machine:"):
		return handleAdminDeleteMachine(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_activate:"):
		return handleAdminActivateMachine(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_deactivate:"):
		return handleAdminDeactivateMachine(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "laundry:admin_bookings":
		return handleAdminViewBookings(ctx, userCtx, svc, fsmMgr)
	case payload == "laundry:admin_create_booking":
		return handleAdminCreateBookingStart(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_booking_machine:"):
		return handleAdminBookingMachine(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_booking_slot:"):
		return handleAdminBookingSlot(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_cancel_booking:"):
		return handleAdminCancelBooking(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "laundry:admin_settings":
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_edit_floor:"):
		return handleAdminEditMachineFloor(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_edit_duration:"):
		return handleAdminEditMachineDuration(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_settings_edit:"):
		return handleAdminLaundrySettingsEdit(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:admin_blocked_machine:"):
		return handleAdminBlockedMachine(ctx, userCtx, payload, svc, fsmMgr)
	default:
		return HandleAdminCallback(ctx, userCtx, payload, svc, fsmMgr)
	}
}

func HandleAdminLaundryMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingMachineData {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)

	switch step {
	case "number":
		userCtx.ContextData["number"] = msg
		userCtx.ContextData["step"] = "duration"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите длительность стирки (в минутах):", keyboards.GetBackKeyboard()), nil

	case "duration":
		duration, err := strconv.Atoi(msg)
		if err != nil || duration < 10 {
			return types.ResponseWithButtons("❌ Неверная длительность. Введите число (минимум 10):", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["duration"] = float64(duration)
		userCtx.ContextData["step"] = "floor"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите номер этажа:", keyboards.GetBackKeyboard()), nil

	case "floor":
		floorNum, err := strconv.Atoi(msg)
		if err != nil || floorNum < 1 {
			return types.ResponseWithButtons("❌ Неверный этаж. Введите число:", keyboards.GetBackKeyboard()), nil
		}

		did, _ := uuid.Parse(userCtx.DormitoryID)
		number := userCtx.ContextData["number"].(string)
		duration := int(userCtx.ContextData["duration"].(float64))

		floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
		var floorID uuid.UUID
		for _, f := range floors {
			if f.FloorNumber == floorNum {
				floorID = f.ID
				break
			}
		}

		_, err = svc.LaundryService.CreateWashingMachine(did, floorID, number, duration)
		if err != nil {
			log.Printf("Error creating washing machine: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)

	case "edit_floor":
		floorNum, err := strconv.Atoi(msg)
		if err != nil || floorNum < 1 {
			return types.ResponseWithButtons("❌ Неверный этаж. Введите число:", keyboards.GetBackKeyboard()), nil
		}

		machineID := userCtx.ContextData["machine_id"].(string)
		mid, _ := uuid.Parse(machineID)
		did, _ := uuid.Parse(userCtx.DormitoryID)
		machines, _ := svc.LaundryService.GetWashingMachines(did)

		for _, m := range machines {
			if m.ID == mid {
				floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
				var floorID uuid.UUID
				for _, f := range floors {
					if f.FloorNumber == floorNum {
						floorID = f.ID
						break
					}
				}
				svc.LaundryService.UpdateWashingMachine(mid, floorID, m.MachineNumber, m.DurationMinutes, m.IsActive)
				break
			}
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)

	case "edit_duration":
		duration, err := strconv.Atoi(msg)
		if err != nil || duration < 10 {
			return types.ResponseWithButtons("❌ Неверная длительность. Введите число (минимум 10):", keyboards.GetBackKeyboard()), nil
		}

		machineID := userCtx.ContextData["machine_id"].(string)
		mid, _ := uuid.Parse(machineID)
		did, _ := uuid.Parse(userCtx.DormitoryID)
		machines, _ := svc.LaundryService.GetWashingMachines(did)

		for _, m := range machines {
			if m.ID == mid {
				svc.LaundryService.UpdateWashingMachine(mid, m.FloorID, m.MachineNumber, duration, m.IsActive)
				break
			}
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)
	}

	return nil, nil
}

func handleAdminLaundry(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateLaundryManagement
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)
	today := timeutil.Today()
	nowStr := timeutil.TimeStr()
	allBookings, _ := svc.LaundryService.GetAllDormitoryBookings(did, today)
	activeOnly, _ := svc.LaundryService.GetDormitoryBookings(did, today)

	floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
	floorMap := make(map[uuid.UUID]int)
	for _, f := range floors {
		floorMap[f.ID] = f.FloorNumber
	}

	bookingByMachine := make(map[uuid.UUID][]domain.LaundryBooking)
	var userIDs []uuid.UUID
	for _, b := range allBookings {
		if b.Status == "active" || b.Status == "completed" {
			bookingByMachine[b.MachineID] = append(bookingByMachine[b.MachineID], b)
			userIDs = append(userIDs, b.UserID)
		}
	}
	usersMap, _ := svc.GetUsersByIDs(userIDs)

	settings, _ := svc.LaundryService.GetLaundrySettings(did)
	bookingStart := "07:00"
	bookingEnd := "23:00"
	if settings != nil {
		if settings.BookingStartTime != "" {
			bookingStart = settings.BookingStartTime
		}
		if settings.BookingEndTime != "" {
			bookingEnd = settings.BookingEndTime
		}
	}

	msg := "👕 Управление прачкой\n\n"
	msg += fmt.Sprintf("Машин: %d | Активных стирок: %d\n", len(machines), len(activeOnly))
	msg += fmt.Sprintf("Запись: %s – %s\n\n", bookingStart, bookingEnd)

	for _, m := range machines {
		bList := bookingByMachine[m.ID]
		status := "✅ свободна"
		if !m.IsActive {
			status = "⛔ выключена"
		} else {
			var prev, current, next *domain.LaundryBooking
			for i := range bList {
				b := &bList[i]
				if b.SlotStart <= nowStr && b.SlotEnd > nowStr {
					current = b
				} else if b.SlotStart > nowStr && (next == nil || b.SlotStart < next.SlotStart) {
					next = b
				} else if b.SlotEnd <= nowStr && (prev == nil || b.SlotEnd > prev.SlotEnd) {
					prev = b
				}
			}

			if current != nil {
				who := formatBookerShort(current.UserID, usersMap)
				status = fmt.Sprintf("🔴 сейчас: %s (%s–%s)", who, current.SlotStart, current.SlotEnd)
			} else if next != nil {
				who := formatBookerShort(next.UserID, usersMap)
				status = fmt.Sprintf("🟡 далее: %s (%s–%s)", who, next.SlotStart, next.SlotEnd)
			}
			if prev != nil && current == nil {
				who := formatBookerShort(prev.UserID, usersMap)
				status += fmt.Sprintf(" | было: %s (%s–%s)", who, prev.SlotStart, prev.SlotEnd)
			}
		}
		if m.BlockedSlotStart != "" {
			status += fmt.Sprintf(" | 👔 каст. %s–%s", m.BlockedSlotStart, m.BlockedSlotEnd)
		}
		msg += fmt.Sprintf("• №%s (этаж %d, %d мин) — %s\n", m.MachineNumber, floorMap[m.FloorID], m.DurationMinutes, status)
	}

	buttons := [][]keyboards.Button{}
	if hasPermission(userCtx, svc, "laundry", "update") {
		buttons = append(buttons, []keyboards.Button{{Text: fmt.Sprintf("🔧 Машины (%d)", len(machines)), Payload: "laundry:admin_machines"}})
		buttons = append(buttons, []keyboards.Button{{Text: "⚙️ Настройки", Payload: "laundry:admin_settings"}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: fmt.Sprintf("📋 Брони (%d)", len(activeOnly)), Payload: "laundry:admin_bookings"}})
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func formatBookerShort(userID uuid.UUID, users map[uuid.UUID]*domain.User) string {
	if u, ok := users[userID]; ok {
		return fmt.Sprintf("%s %s", u.LastName, u.FirstName)
	}
	return userID.String()[:8]
}

func handleAdminListMachines(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	did, _ := uuid.Parse(userCtx.DormitoryID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
	floorMap := make(map[uuid.UUID]int)
	for _, f := range floors {
		floorMap[f.ID] = f.FloorNumber
	}

	msg := "🔧 Стиральные машины:\n\n"

	if len(machines) == 0 {
		msg += "Машин пока нет.\n"
	}

	buttons := [][]keyboards.Button{}
	for _, m := range machines {
		status := "✅ активна"
		if !m.IsActive {
			status = "⛔ выключена"
		}
		msg += fmt.Sprintf("• №%s (этаж %d, %d мин) — %s\n", m.MachineNumber, floorMap[m.FloorID], m.DurationMinutes, status)

		if !hasPermission(userCtx, svc, "laundry", "update") {
			continue
		}

		row := []keyboards.Button{{
			Text:    fmt.Sprintf("⚙️ %s", m.MachineNumber),
			Payload: fmt.Sprintf("laundry:admin_edit_machine:%s", m.ID.String()),
		}}
		if m.IsActive {
			row = append(row, keyboards.Button{
				Text:    "⛔ Выкл",
				Payload: fmt.Sprintf("laundry:admin_deactivate:%s", m.ID.String()),
			})
		} else {
			row = append(row, keyboards.Button{
				Text:    "✅ Вкл",
				Payload: fmt.Sprintf("laundry:admin_activate:%s", m.ID.String()),
			})
		}
		buttons = append(buttons, row)
	}

	if hasPermission(userCtx, svc, "laundry", "create") {
		buttons = append(buttons, []keyboards.Button{{Text: "➕ Добавить", Payload: "laundry:admin_create_machine"}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleAdminCreateMachineStart(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "create"); !ok {
		return blocked, nil
	}
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingMachineData
	userCtx.ContextData = map[string]interface{}{"step": "number"}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите номер машины:", keyboards.GetBackKeyboard()), nil
}

func handleAdminEditMachineFloor(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	machineID := strings.TrimPrefix(payload, "laundry:admin_edit_floor:")
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingMachineData
	userCtx.ContextData = map[string]interface{}{"step": "edit_floor", "machine_id": machineID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите новый номер этажа:", keyboards.GetBackKeyboard()), nil
}

func handleAdminEditMachineDuration(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	machineID := strings.TrimPrefix(payload, "laundry:admin_edit_duration:")
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingMachineData
	userCtx.ContextData = map[string]interface{}{"step": "edit_duration", "machine_id": machineID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите новую длительность (мин):", keyboards.GetBackKeyboard()), nil
}

func handleAdminEditMachineStart(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "update"); !ok {
		return blocked, nil
	}
	machineID := strings.TrimPrefix(payload, "laundry:admin_edit_machine:")
	mid, _ := uuid.Parse(machineID)

	did, _ := uuid.Parse(userCtx.DormitoryID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	var target *domain.WashingMachine
	for _, m := range machines {
		if m.ID == mid {
			target = &m
			break
		}
	}
	if target == nil {
		return types.ResponseWithButtons("❌ Машина не найдена.", keyboards.GetBackKeyboard()), nil
	}

	floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
	floorMap := make(map[uuid.UUID]int)
	for _, f := range floors {
		floorMap[f.ID] = f.FloorNumber
	}
	floorNum := floorMap[target.FloorID]

	statusText := "Выключена"
	if target.IsActive {
		statusText = "Включена"
	}
	msg := fmt.Sprintf("🔧 Машина №%s\nЭтаж: %d\nДлительность: %d мин\nСтатус: %s\n\nВыберите действие:",
		target.MachineNumber, floorNum, target.DurationMinutes, statusText)

	return types.ResponseWithButtons(msg, [][]keyboards.Button{
		{{Text: "🏢 Изменить этаж", Payload: fmt.Sprintf("laundry:admin_edit_floor:%s", target.ID.String())}},
		{{Text: "⏱ Изменить длительность", Payload: fmt.Sprintf("laundry:admin_edit_duration:%s", target.ID.String())}},
		{{Text: "🗑️ Удалить", Payload: fmt.Sprintf("laundry:admin_delete_machine:%s", target.ID.String())}},
		{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
	}), nil
}

func handleAdminDeleteMachine(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "delete"); !ok {
		return blocked, nil
	}
	machineID := strings.TrimPrefix(payload, "laundry:admin_delete_machine:")
	mid, _ := uuid.Parse(machineID)
	if err := svc.LaundryService.DeleteWashingMachine(mid); err != nil {
		log.Printf("Error deleting washing machine: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)
}

func handleAdminActivateMachine(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "update"); !ok {
		return blocked, nil
	}
	machineID := strings.TrimPrefix(payload, "laundry:admin_activate:")
	mid, _ := uuid.Parse(machineID)

	did, _ := uuid.Parse(userCtx.DormitoryID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)
	for _, m := range machines {
		if m.ID == mid {
			svc.LaundryService.UpdateWashingMachine(mid, m.FloorID, m.MachineNumber, m.DurationMinutes, true)
			break
		}
	}
	return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)
}

func handleAdminDeactivateMachine(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "update"); !ok {
		return blocked, nil
	}
	machineID := strings.TrimPrefix(payload, "laundry:admin_deactivate:")
	mid, _ := uuid.Parse(machineID)

	did, _ := uuid.Parse(userCtx.DormitoryID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)
	for _, m := range machines {
		if m.ID == mid {
			svc.LaundryService.UpdateWashingMachine(mid, m.FloorID, m.MachineNumber, m.DurationMinutes, false)
			break
		}
	}
	return handleAdminListMachines(ctx, userCtx, svc, fsmMgr)
}

func handleAdminViewBookings(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	did, _ := uuid.Parse(userCtx.DormitoryID)
	today := timeutil.Today()
	bookings, _ := svc.LaundryService.GetDormitoryBookings(did, today)
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	machineMap := make(map[uuid.UUID]string)
	for _, m := range machines {
		machineMap[m.ID] = m.MachineNumber
	}

	var userIDs []uuid.UUID
	for _, b := range bookings {
		if b.Status == "active" {
			userIDs = append(userIDs, b.UserID)
		}
	}
	usersMap, _ := svc.GetUsersByIDs(userIDs)

	msg := fmt.Sprintf("📋 Брони на %s:\n\n", timeutil.FormatEU(today))

	if len(bookings) == 0 {
		msg += "Броней нет.\n"
		return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil
	}

	buttons := [][]keyboards.Button{}
	for _, b := range bookings {
		if b.Status != "active" {
			continue
		}
		machineNum := machineMap[b.MachineID]
		if machineNum == "" {
			machineNum = b.MachineID.String()[:8]
		}
		residentLabel := b.UserID.String()[:8]
		if u, ok := usersMap[b.UserID]; ok {
			residentLabel = fmt.Sprintf("%s %s", u.LastName, u.FirstName)
		}
		msg += fmt.Sprintf("• №%s: %s–%s (%s)\n", machineNum, b.SlotStart, b.SlotEnd, residentLabel)
		if hasPermission(userCtx, svc, "laundry", "delete") {
			buttons = append(buttons, []keyboards.Button{{
				Text:    fmt.Sprintf("❌ Отменить №%s %s–%s", machineNum, b.SlotStart, b.SlotEnd),
				Payload: fmt.Sprintf("laundry:admin_cancel_booking:%s", b.ID.String()),
			}})
		}
	}

	if hasPermission(userCtx, svc, "laundry", "create") {
		buttons = append(buttons, []keyboards.Button{{Text: "➕ Создать бронь", Payload: "laundry:admin_create_booking"}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleAdminCancelBooking(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "delete"); !ok {
		return blocked, nil
	}
	bookingID := strings.TrimPrefix(payload, "laundry:admin_cancel_booking:")
	bid, _ := uuid.Parse(bookingID)

	svc.NotificationService.CancelRemindersForBooking(ctx, bid.String())

	if err := svc.LaundryService.AdminCancelBooking(bid); err != nil {
		log.Printf("Error admin cancelling booking: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return handleAdminViewBookings(ctx, userCtx, svc, fsmMgr)
}

func handleAdminLaundrySettings(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if !hasPermission(userCtx, svc, "laundry", "update") {
		return handleAdminLaundrySettingsView(ctx, userCtx, svc)
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	settings, err := svc.LaundryService.GetLaundrySettings(did)

	if err != nil || settings == nil {
		settings = &domain.LaundrySettings{
			DormitoryID:            did,
			BookingStartTime:       "07:00",
			DefaultDurationMinutes: 60,
		}
	}

	machines, _ := svc.LaundryService.GetWashingMachines(did)
	msg := "⚙️ Настройки прачечной\n\n"
	msg += fmt.Sprintf("Начало записи: %s\n", settings.BookingStartTime)
	msg += fmt.Sprintf("Конец записи: %s\n", settings.BookingEndTime)
	msg += fmt.Sprintf("Длительность слота: %d мин\n", settings.DefaultDurationMinutes)
	advanceText := "Выкл"
	if settings.AdvanceBookingEnabled {
		advanceText = fmt.Sprintf("Вкл (до %d дн.)", settings.AdvanceBookingMaxDays)
	}
	msg += fmt.Sprintf("Запись заранее: %s\n", advanceText)
	castellanText := "Вкл"
	if !settings.CastellanBookingEnabled {
		castellanText = "Выкл"
	}
	msg += fmt.Sprintf("Бронь кастелянши: %s\n", castellanText)

	hasBlocked := false
	for _, m := range machines {
		if m.BlockedSlotStart != "" {
			if !hasBlocked {
				msg += "\nСлоты кастелянши:\n"
				hasBlocked = true
			}
			msg += fmt.Sprintf("• Машина %s: %s–%s (%s)\n", m.MachineNumber, m.BlockedSlotStart, m.BlockedSlotEnd, m.BlockedSlotReason)
		}
	}
	if !hasBlocked {
		msg += "\nСлот кастелянши: не задан\n"
	}

	return types.ResponseWithButtons(msg, [][]keyboards.Button{
		{{Text: "🕐 Начало записи", Payload: "laundry:admin_settings_edit:start_time"}},
		{{Text: "🕐 Конец записи", Payload: "laundry:admin_settings_edit:end_time"}},
		{{Text: "⏱ Длительность", Payload: "laundry:admin_settings_edit:duration"}},
		{{Text: "📅 Запись заранее", Payload: "laundry:admin_settings_edit:advance_booking"}},
		{{Text: "📆 Дней вперёд", Payload: "laundry:admin_settings_edit:advance_days"}},
		{{Text: "👔 Слот кастелянши", Payload: "laundry:admin_settings_edit:blocked"}},
		{{Text: "👔 Бронь кастелянши", Payload: "laundry:admin_settings_edit:castellan_toggle"}},
		{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}},
	}), nil
}

func handleAdminLaundrySettingsView(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService) (*types.MessageResponse, error) {
	did, _ := uuid.Parse(userCtx.DormitoryID)
	settings, _ := svc.LaundryService.GetLaundrySettings(did)
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	msg := "⚙️ Настройки прачечной\n\n"
	if settings == nil {
		msg += "Начало записи: 07:00\nДлительность слота: 60 мин\nЗапись заранее: Выкл\n"
	} else {
		msg += fmt.Sprintf("Начало записи: %s\n", settings.BookingStartTime)
		msg += fmt.Sprintf("Длительность слота: %d мин\n", settings.DefaultDurationMinutes)
		advanceText := "Выкл"
		if settings.AdvanceBookingEnabled {
			advanceText = fmt.Sprintf("Вкл (до %d дн.)", settings.AdvanceBookingMaxDays)
		}
		msg += fmt.Sprintf("Запись заранее: %s\n", advanceText)	}

	for _, m := range machines {
		if m.BlockedSlotStart != "" {
			msg += fmt.Sprintf("\nМашина %s (кастелянша): %s–%s", m.MachineNumber, m.BlockedSlotStart, m.BlockedSlotEnd)
		}
	}
	return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil
}

func handleAdminLaundrySettingsEdit(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	field := strings.TrimPrefix(payload, "laundry:admin_settings_edit:")

	switch field {
	case "start_time":
		userCtx.PushState()
		userCtx.State = fsm.StateAwaitingLaundrySettings
		userCtx.ContextData = map[string]interface{}{"step": "start_time"}
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите время начала записи (ЧЧ:ММ):", keyboards.GetBackKeyboard()), nil

	case "end_time":
		userCtx.PushState()
		userCtx.State = fsm.StateAwaitingLaundrySettings
		userCtx.ContextData = map[string]interface{}{"step": "end_time"}
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите время окончания записи (ЧЧ:ММ):", keyboards.GetBackKeyboard()), nil

	case "duration":
		userCtx.PushState()
		userCtx.State = fsm.StateAwaitingLaundrySettings
		userCtx.ContextData = map[string]interface{}{"step": "duration"}
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите длительность слота (минут):", keyboards.GetBackKeyboard()), nil

	case "advance_booking":
		did, _ := uuid.Parse(userCtx.DormitoryID)
		settings, _ := svc.LaundryService.GetLaundrySettings(did)
		if settings == nil {
			settings = &domain.LaundrySettings{DormitoryID: did, BookingStartTime: "07:00", DefaultDurationMinutes: 60}
		}
		settings.AdvanceBookingEnabled = !settings.AdvanceBookingEnabled
		if err := svc.LaundryService.UpdateLaundrySettings(settings); err != nil {
			log.Printf("Error toggling advance booking: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)

	case "advance_days":
		userCtx.PushState()
		userCtx.State = fsm.StateAwaitingLaundrySettings
		userCtx.ContextData = map[string]interface{}{"step": "advance_days"}
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите количество дней для предварительной записи (1–14):", keyboards.GetBackKeyboard()), nil

	case "castellan_toggle":
		did, _ := uuid.Parse(userCtx.DormitoryID)
		settings, _ := svc.LaundryService.GetLaundrySettings(did)
		if settings == nil {
			settings = &domain.LaundrySettings{DormitoryID: did, BookingStartTime: "07:00", DefaultDurationMinutes: 60}
		}
		settings.CastellanBookingEnabled = !settings.CastellanBookingEnabled
		if err := svc.LaundryService.UpdateLaundrySettings(settings); err != nil {
			log.Printf("Error toggling castellan booking: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		enabledText := "Выкл"
		if settings.CastellanBookingEnabled {
			enabledText = "Вкл"
		}
		msg := fmt.Sprintf("👔 Бронь кастелянши: %s\n\nФункция позволяет кастелянше резервировать слоты для служебной стирки (постельное бельё, полотенца).\nЖители не могут записаться на эти слоты.", enabledText)
		return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil

	case "blocked":
		did, _ := uuid.Parse(userCtx.DormitoryID)
		machines, _ := svc.LaundryService.GetWashingMachines(did)
		btns := [][]keyboards.Button{}
		for _, m := range machines {
			btns = append(btns, []keyboards.Button{{
				Text:    fmt.Sprintf("Машина %s", m.MachineNumber),
				Payload: fmt.Sprintf("laundry:admin_blocked_machine:%s", m.ID.String()),
			}})
		}
		btns = append(btns, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
		return types.ResponseWithButtons("Выберите машину для настройки слота кастелянши:", btns), nil
	}

	return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)
}

func handleAdminBlockedMachine(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	machineID := strings.TrimPrefix(payload, "laundry:admin_blocked_machine:")
	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingLaundrySettings
	userCtx.ContextData = map[string]interface{}{"step": "blocked_start", "blocked_machine_id": machineID}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите время начала слота кастелянши (ЧЧ:ММ) или 0 для отключения:", keyboards.GetBackKeyboard()), nil
}

func HandleAdminLaundrySettingsMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingLaundrySettings {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)
	did, _ := uuid.Parse(userCtx.DormitoryID)

	settings, _ := svc.LaundryService.GetLaundrySettings(did)
	if settings == nil {
		settings = &domain.LaundrySettings{DormitoryID: did, BookingStartTime: "07:00", DefaultDurationMinutes: 60}
	}

	switch step {
	case "start_time":
		if !strings.Contains(msg, ":") || len(msg) != 5 {
			return types.ResponseWithButtons("❌ Неверный формат. Введите время как ЧЧ:ММ:", keyboards.GetBackKeyboard()), nil
		}
		settings.BookingStartTime = msg
		if err := svc.LaundryService.UpdateLaundrySettings(settings); err != nil {
			log.Printf("Error updating start time: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)

	case "end_time":
		if !strings.Contains(msg, ":") || len(msg) != 5 {
			return types.ResponseWithButtons("❌ Неверный формат. Введите время как ЧЧ:ММ:", keyboards.GetBackKeyboard()), nil
		}
		settings.BookingEndTime = msg
		if err := svc.LaundryService.UpdateLaundrySettings(settings); err != nil {
			log.Printf("Error updating end time: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)

	case "duration":
		dur, err := strconv.Atoi(msg)
		if err != nil || dur < 10 {
			return types.ResponseWithButtons("❌ Неверная длительность. Введите число (минимум 10):", keyboards.GetBackKeyboard()), nil
		}
		settings.DefaultDurationMinutes = dur
		if err := svc.LaundryService.UpdateLaundrySettings(settings); err != nil {
			log.Printf("Error updating duration: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		svc.LaundryService.UpdateAllMachinesDuration(did, dur)
		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)

	case "advance_days":
		days, err := strconv.Atoi(msg)
		if err != nil || days < 1 || days > 14 {
			return types.ResponseWithButtons("❌ Введите число от 1 до 14:", keyboards.GetBackKeyboard()), nil
		}
		settings.AdvanceBookingMaxDays = days
		if err := svc.LaundryService.UpdateLaundrySettings(settings); err != nil {
			log.Printf("Error updating advance days: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)

	case "blocked_start":
		machineID := userCtx.ContextData["blocked_machine_id"].(string)
		mid, _ := uuid.Parse(machineID)

		if msg == "0" {
			if err := svc.LaundryService.SetWashingMachineBlockedSlot(mid, "", "", ""); err != nil {
				log.Printf("Error clearing blocked slot: %v", err)
				return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
			}
			userCtx.ContextData = nil
			userCtx.State = fsm.StateLaundryManagement
			fsmMgr.Set(ctx, userCtx)
			return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)
		}
		if !strings.Contains(msg, ":") || len(msg) != 5 {
			return types.ResponseWithButtons("❌ Неверный формат. Введите ЧЧ:ММ:", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["blocked_start"] = msg
		userCtx.ContextData["step"] = "blocked_end"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите время окончания (ЧЧ:ММ):", keyboards.GetBackKeyboard()), nil

	case "blocked_end":
		if !strings.Contains(msg, ":") || len(msg) != 5 {
			return types.ResponseWithButtons("❌ Неверный формат. Введите ЧЧ:ММ:", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["blocked_end"] = msg
		userCtx.ContextData["step"] = "blocked_reason"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите причину блокировки (например, «Стирка кастелянши»):", keyboards.GetBackKeyboard()), nil

	case "blocked_reason":
		machineID := userCtx.ContextData["blocked_machine_id"].(string)
		mid, _ := uuid.Parse(machineID)
		blockedStart := userCtx.ContextData["blocked_start"].(string)
		blockedEnd := userCtx.ContextData["blocked_end"].(string)
		blockedReason := msg
		if err := svc.LaundryService.SetWashingMachineBlockedSlot(mid, blockedStart, blockedEnd, blockedReason); err != nil {
			log.Printf("Error setting blocked slot: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData = nil
		userCtx.State = fsm.StateLaundryManagement
		fsmMgr.Set(ctx, userCtx)
		return handleAdminLaundrySettings(ctx, userCtx, svc, fsmMgr)
	}

	return nil, nil
}

func handleAdminCreateBookingStart(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "laundry", "create"); !ok {
		return blocked, nil
	}
	userCtx.State = fsm.StateAwaitingAdminBooking
	userCtx.ContextData = map[string]interface{}{"step": "phone"}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите номер телефона жильца (10 цифр):", keyboards.GetBackKeyboard()), nil
}

func HandleAdminBookingMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingAdminBooking {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)
	did, _ := uuid.Parse(userCtx.DormitoryID)

	switch step {
	case "phone":
		resident, phone, err := lookupResidentByPhone(ctx, svc, msg)
		if err != nil || resident == nil {
			return types.ResponseWithButtons("❌ Жилец с таким телефоном не найден.", keyboards.GetBackKeyboard()), nil
		}
		userCtx.ContextData["user_id"] = resident.UserID.String()
		userCtx.ContextData["resident_id"] = resident.ID.String()
		userCtx.ContextData["resident_name"] = fmt.Sprintf("%s %s", resident.User.LastName, resident.User.FirstName)
		puID, _ := strconv.ParseInt(resident.User.PlatformUserID, 10, 64)
		userCtx.ContextData["platform_user_id"] = puID

		machines, _ := svc.LaundryService.GetWashingMachines(did)
		if len(machines) == 0 {
			return types.ResponseWithButtons("❌ Нет машин.", keyboards.GetBackKeyboard()), nil
		}

		resp := fmt.Sprintf("Жилец: %s (%s)\n\nВыберите машину:", resident.User.LastName+" "+resident.User.FirstName, phone)
		btns := [][]keyboards.Button{}
		for _, m := range machines {
			if m.IsActive {
				btns = append(btns, []keyboards.Button{{
					Text:    fmt.Sprintf("Машина %s", m.MachineNumber),
					Payload: fmt.Sprintf("laundry:admin_booking_machine:%s", m.ID.String()),
				}})
			}
		}
		btns = append(btns, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
		return types.ResponseWithButtons(resp, btns), nil
	}

	return nil, nil
}

func handleAdminBookingMachine(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	base := keyboards.StripPageSuffix(payload)
	page := keyboards.ParsePage(payload)
	machineID := strings.TrimPrefix(base, "laundry:admin_booking_machine:")

	did, _ := uuid.Parse(userCtx.DormitoryID)
	mid, _ := uuid.Parse(machineID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	var target *domain.WashingMachine
	for _, m := range machines {
		if m.ID == mid {
			target = &m
			break
		}
	}
	if target == nil {
		return types.ResponseWithButtons("❌ Машина не найдена.", keyboards.GetBackKeyboard()), nil
	}

	duration := target.DurationMinutes
	if duration == 0 {
		duration = 60
	}

	settings, _ := svc.LaundryService.GetLaundrySettings(did)
	var startTime string
	if settings != nil && settings.BookingStartTime != "" {
		startTime = settings.BookingStartTime
	} else {
		startTime = "07:00"
	}

	endTimeStr := "23:00"
	if settings != nil && settings.BookingEndTime != "" {
		endTimeStr = settings.BookingEndTime
	}

	today := timeutil.Today()

	var defaultStartHour, defaultStartMin int
	fmt.Sscanf(startTime, "%d:%d", &defaultStartHour, &defaultStartMin)
	var endHour, endMin int
	fmt.Sscanf(endTimeStr, "%d:%d", &endHour, &endMin)
	slotStart := time.Date(0, 1, 1, defaultStartHour, defaultStartMin, 0, 0, time.UTC)
	slotEndTime := time.Date(0, 1, 1, endHour, endMin, 0, 0, time.UTC)

	allBookings, _ := svc.LaundryService.GetDormitoryBookings(did, today)

	var items []keyboards.PageableButton
	now := timeutil.Now()

	for slotStart.Before(slotEndTime) {
		slotEnd := slotStart.Add(time.Duration(duration) * time.Minute)
		if slotEnd.After(slotEndTime) {
			break
		}

		startStr := slotStart.Format("15:04")
		endStr := slotEnd.Format("15:04")

		isPast := startStr <= now.Format("15:04")

		conflict := false
		for _, b := range allBookings {
			if b.MachineID == mid && b.SlotStart < endStr && b.SlotEnd > startStr && b.Status == "active" {
				conflict = true
				break
			}
		}

		isBlocked := target.BlockedSlotStart != "" && startStr < target.BlockedSlotEnd && endStr > target.BlockedSlotStart

		if isBlocked {
			items = append(items, keyboards.PageableButton{
				Text:    fmt.Sprintf("👔 %s–%s (кастелянша)", startStr, endStr),
				Payload: "none",
			})
		} else if !conflict && !isPast {
			items = append(items, keyboards.PageableButton{
				Text:    fmt.Sprintf("✅ %s–%s", startStr, endStr),
				Payload: fmt.Sprintf("laundry:admin_booking_slot:%s:%s", machineID, startStr),
			})
		} else if !isPast {
			items = append(items, keyboards.PageableButton{
				Text:    fmt.Sprintf("🔴 %s–%s (занято)", startStr, endStr),
				Payload: "none",
			})
		}

		slotStart = slotEnd
	}

	prefix := fmt.Sprintf("laundry:admin_booking_machine:%s", machineID)
	buttons := keyboards.BuildPaginatedKeyboard(items, prefix, page, 8)

	return types.ResponseWithButtons(fmt.Sprintf("Выберите слот для машины %s:", target.MachineNumber), buttons), nil
}

func handleAdminBookingSlot(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	payload = strings.TrimPrefix(payload, "laundry:admin_booking_slot:")
	idx := strings.Index(payload, ":")
	if idx < 0 {
		return types.ResponseWithButtons("❌ Ошибка выбора слота.", keyboards.GetBackKeyboard()), nil
	}
	machineID, slotStart := payload[:idx], payload[idx+1:]

	mid, _ := uuid.Parse(machineID)
	uid, _ := uuid.Parse(userCtx.ContextData["user_id"].(string))
	did, _ := uuid.Parse(userCtx.DormitoryID)
	today := timeutil.Today()

	machines, _ := svc.LaundryService.GetWashingMachines(did)
	var duration int
	var machineNumber string
	for _, m := range machines {
		if m.ID == mid {
			duration = m.DurationMinutes
			machineNumber = m.MachineNumber
			break
		}
	}
	if duration == 0 {
		duration = 60
	}

	slotEnd := calculateSlotEnd(slotStart, duration)

	booking, err := svc.LaundryService.CreateBooking(uid, "resident", mid, today, slotStart, slotEnd)
	if err != nil {
		log.Printf("Error creating admin booking: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	puID, _ := userCtx.ContextData["platform_user_id"].(int64)
	if puID > 0 {
		svc.NotificationService.ScheduleBookingReminders(ctx, machineNumber, today, booking.SlotStart, booking.SlotEnd, puID)
	}

	userCtx.ContextData = nil
	userCtx.State = fsm.StateLaundryManagement
	fsmMgr.Set(ctx, userCtx)

	msg := fmt.Sprintf("✅ Бронь создана.\nМашина №%s\n%s %s–%s",
		machineNumber, timeutil.FormatEU(today), booking.SlotStart, booking.SlotEnd)
	return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil
}

func lookupResidentByPhone(_ context.Context, svc *service.DBService, phone string) (*domain.Resident, string, error) {
	phone = strings.TrimSpace(phone)
	if len(phone) == 10 {
		phone = "7" + phone
	}
	if len(phone) > 11 {
		phone = phone[len(phone)-11:]
	}

	user, err := svc.GetUserByPhone(phone)
	if err != nil || user == nil {
		return nil, phone, fmt.Errorf("пользователь не найден")
	}

	resident, err := svc.GetResidentByUserID(user.ID)
	if err != nil || resident == nil {
		return nil, phone, fmt.Errorf("жилец не найден")
	}

	return resident, phone, nil
}
