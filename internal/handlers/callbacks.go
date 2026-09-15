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

func HandleCallbackQuery(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	log.Printf("User %d: Callback payload: %s (state: %s)", userCtx.UserID, payload, userCtx.State)

	if keyboards.HasPageSuffix(payload) {
		base := keyboards.StripPageSuffix(payload)
		page := keyboards.ParsePage(payload)
		switch {
		case strings.HasPrefix(base, "info:category:list"):
			return handleMenuInfo(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "info:list"):
			return handleMenuInfo(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "room:items:"):
			return handleMenuRoom(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "laundry:slots"):
			return handleMenuLaundry(ctx, userCtx, svc, fsmMgr, page)
		case strings.HasPrefix(base, "laundry:machine:"):
			return handleLaundryMachineSlots(ctx, userCtx, base, svc, fsmMgr, page)
		case strings.HasPrefix(base, "module_refs:list"):
			return handleModuleRefsPaged(ctx, userCtx, svc, fsmMgr, page)
		}
	}

	switch {
	case payload == "auth:request":
		return handleAuthRequest(ctx, userCtx, fsmMgr)
	case payload == keyboards.PayloadResidentConfirm:
		return handleResidentConfirm(ctx, userCtx, svc, fsmMgr)
	case payload == keyboards.PayloadResidentReject:
		return handleResidentReject(ctx, userCtx, fsmMgr)
	case payload == "role:resident":
		return handleRoleResident(ctx, userCtx, svc, fsmMgr)
	case payload == "role:employee":
		return handleRoleEmployee(ctx, userCtx, svc, fsmMgr)
	case payload == keyboards.PayloadMenuLaundry:
		return handleMenuLaundry(ctx, userCtx, svc, fsmMgr, 0)
	case payload == keyboards.PayloadMenuRoom:
		return handleMenuRoom(ctx, userCtx, svc, fsmMgr, 0)
	case payload == keyboards.PayloadMenuReferences:
		return handleMenuInfo(ctx, userCtx, svc, fsmMgr, 0)
	case payload == keyboards.PayloadMenuChatLinks:
		return handleMenuChatLinks(ctx, userCtx, svc, fsmMgr)
	case payload == keyboards.PayloadMenuCleaning:
		return handleMenuCleaning(ctx, userCtx, svc, fsmMgr)
	case payload == keyboards.PayloadNavBack:
		return handleNavBack(ctx, userCtx, svc, fsmMgr)
	case payload == keyboards.PayloadNavHome:
		return handleNavHome(ctx, userCtx, svc, fsmMgr)
	case payload == keyboards.PayloadLaundryViewQueue:
		return handleLaundryViewQueue(ctx, userCtx, svc, fsmMgr, 0)
	case payload == keyboards.PayloadRoomViewItems:
		return handleRoomViewItems(ctx, userCtx, svc, fsmMgr, 0)
	case payload == keyboards.PayloadInfoViewList:
		return handleInfoViewList(ctx, userCtx, svc, fsmMgr, 0)
	case payload == keyboards.PayloadCleaningView:
		return handleCleaningView(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:book:"):
		return handleLaundryBook(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:cancel_confirm:"):
		return handleLaundryCancelConfirm(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:cancel:"):
		return handleLaundryCancel(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:slots:"):
		return handleLaundrySlots(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:machine:"):
		return handleLaundryMachineSlots(ctx, userCtx, payload, svc, fsmMgr, 0)
	case strings.HasPrefix(payload, "laundry:my_bookings"):
		return handleLaundryMyBookings(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "laundry:date:"):
		return handleLaundrySlotsDate(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "info:view:"):
		return handleInfoViewItem(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "info:category:"):
		return handleInfoByCategory(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "module_refs:"):
		return handleModuleRefs(ctx, userCtx, payload, svc, fsmMgr, 0)
	case strings.HasPrefix(payload, "module_ref_view:"):
		return handleModuleRefView(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:resident_complete:"):
		return handleCleaningResidentComplete(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "cleaning:resident_month:"):
		return handleCleaningResidentMonth(ctx, userCtx, payload, svc, fsmMgr)
	case strings.HasPrefix(payload, "dormitory:list"):
		return handleDormitoryList(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "dormitory:select:"):
		return handleDormitorySelect(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "resident_role:select":
		return handleResidentRoleSelect(ctx, userCtx, svc, fsmMgr)
	case strings.HasPrefix(payload, "resident_role:confirm:"):
		return handleResidentRoleConfirm(ctx, userCtx, payload, svc, fsmMgr)
	case payload == "resident_role:exit":
		return handleResidentRoleExit(ctx, userCtx, svc, fsmMgr)
	default:
		log.Printf("Unknown callback payload: %s", payload)
		return types.TextResponse(templates.MsgUnknownCommand), nil
	}
}

func handleInfoByCategory(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	category := strings.TrimPrefix(payload, "info:category:")
	did, _ := uuid.Parse(userCtx.DormitoryID)
	materials, err := svc.ReferenceService.GetByCategory(did, category)
	if err != nil || len(materials) == 0 {
		return types.ResponseWithButtons("📖 В этой категории пока нет материалов.", keyboards.GetBackKeyboard()), nil
	}

	var buttons [][]keyboards.Button
	for _, m := range materials {
		buttons = append(buttons, []keyboards.Button{{
			Text:    m.Name,
			Payload: keyboards.PayloadInfoViewItem(m.ID.String()),
		}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(fmt.Sprintf("📖 %s:", category), buttons), nil
}

func handleAuthRequest(ctx context.Context, userCtx *fsm.UserContext, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.State = fsm.StateAwaitingContact
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons(templates.MsgAskContact, keyboards.GetContactKeyboard()), nil
}

func handleResidentConfirm(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if err := svc.ConfirmResident(userCtx.ResidentUUID); err != nil {
		log.Printf("Error confirming resident: %v", err)
		return nil, err
	}

	userCtx.State = fsm.StateAuthorized
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons(templates.MsgResidentConfirmed, buildResidentMainMenu(ctx, userCtx, svc)), nil
}

func handleResidentReject(ctx context.Context, userCtx *fsm.UserContext, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.State = fsm.StateAwaitingContact
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons(templates.MsgAskContact, keyboards.GetContactKeyboard()), nil
}

func handleRoleResident(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	resident, err := svc.InitResident(userCtx.UserUUID)
	if err != nil || resident == nil {
		return types.TextResponse(templates.MsgResidentNotFound), nil
	}

	userCtx.State = fsm.StateResidentConfirmData
	userCtx.DormitoryID = resident.DormitoryID
	userCtx.DormitoryName = resident.DormitoryName
	userCtx.RoomID = resident.RoomID
	userCtx.RoomNumber = resident.RoomNumber
	userCtx.FloorNumber = resident.FloorNumber
	userCtx.ResidentUUID = resident.ResidentID

	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	msg := templates.FormatResidentData(resident.DormitoryName, resident.RoomNumber, resident.FloorNumber)
	return types.ResponseWithButtons(msg, keyboards.GetConfirmationKeyboard(keyboards.PayloadResidentConfirm, keyboards.PayloadResidentReject)), nil
}

func handleRoleEmployee(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	employee, err := svc.InitEmployee(userCtx.UserUUID)
	if err != nil || employee == nil {
		return types.TextResponse(templates.MsgEmployeeNotFound), nil
	}

	userCtx.EmployeeUUID = employee.EmployeeID
	userCtx.IsAuthorized = true

	dormitories, err := svc.GetEmployeeDormitories(userCtx.EmployeeUUID)
	if err != nil || len(dormitories) == 0 {
		return types.TextResponse("У вас нет доступа ни к одному общежитию."), nil
	}

	if len(dormitories) == 1 {
		d := dormitories[0]
		role, _ := svc.GetEmployeeRoleInDormitory(userCtx.EmployeeUUID, d.DormitoryID)
		rolenames := buildRoleDisplayLookup(svc)
		roleDisplay := service.MapRoleDisplayWithLookup(role, rolenames)
		userCtx.DormitoryID = d.DormitoryID
		userCtx.EmployeeRole = role
		userCtx.State = fsm.StateRoleVerified
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}

		msg := fmt.Sprintf("✅ %s, ваша должность — %s.\nОбщежитие: %s", userCtx.FirstName, roleDisplay, d.DormitoryName)
		return types.ResponseWithButtons(msg, newAdminMenuBuilderFromContext(svc, userCtx).build(ctx)), nil
	}

	userCtx.State = fsm.StateSelectingDormitory
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	var list []struct {
		ID   string
		Name string
	}
	for _, d := range dormitories {
		list = append(list, struct {
			ID   string
			Name string
		}{ID: d.DormitoryID, Name: d.DormitoryName})
	}
	return types.ResponseWithButtons(templates.MsgSelectDormitory, keyboards.GetDormitorySelectKeyboard(list)), nil
}

func handleDormitoryList(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	dormitories, err := svc.GetEmployeeDormitories(userCtx.EmployeeUUID)
	if err != nil || len(dormitories) == 0 {
		return types.TextResponse("Нет доступных общежитий."), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateSelectingDormitory
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	var list []struct {
		ID   string
		Name string
	}
	for _, d := range dormitories {
		list = append(list, struct {
			ID   string
			Name string
		}{ID: d.DormitoryID, Name: d.DormitoryName})
	}
	return types.ResponseWithButtons(templates.MsgSelectDormitory, keyboards.GetDormitorySelectKeyboard(list)), nil
}

func handleDormitorySelect(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	dormitoryID := strings.TrimPrefix(payload, "dormitory:select:")

	role, err := svc.GetEmployeeRoleInDormitory(userCtx.EmployeeUUID, dormitoryID)
	if err != nil {
		return types.TextResponse("Ошибка при определении роли."), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateRoleVerified
	userCtx.DormitoryID = dormitoryID
	userCtx.EmployeeRole = role
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	rolenames := buildRoleDisplayLookup(svc)
	roleDisplay := service.MapRoleDisplayWithLookup(role, rolenames)
	return types.ResponseWithButtons(
		templates.FormatAuthorizedEmployee(userCtx.FirstName, roleDisplay),
		newAdminMenuBuilderFromContext(svc, userCtx).build(ctx),
	), nil
}

func handleMenuLaundry(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if page == 0 {
		userCtx.PushState()
		userCtx.State = fsm.StateLaundry
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	uid, _ := uuid.Parse(userCtx.UserUUID)
	bookerType := "resident"
	if userCtx.EmployeeUUID != "" {
		bookerType = "employee"
	}
	userCtx.ContextData["laundry_booker_type"] = bookerType

	machines, _ := svc.LaundryService.GetWashingMachines(did)

	msg := "👕 Прачка\n\n"
	msg += fmt.Sprintf("Машин в общежитии: %d\n", len(machines))
	today := timeutil.Today()
	active, _ := svc.LaundryService.GetDormitoryBookings(did, today)
	msg += fmt.Sprintf("Активных стирок сегодня: %d\n", len(active))

	myBookings, _ := svc.LaundryService.GetUserBookings(uid)
	hasMy := false
	for _, b := range myBookings {
		if b.Status == "active" && b.SlotDate >= today {
			hasMy = true
			break
		}
	}

	buttons := [][]keyboards.Button{
		{{Text: "📅 Мои брони", Payload: "laundry:my_bookings"}},
		{{Text: "🕐 Свободные слоты", Payload: "laundry:slots:today"}},
	}

	if hasMy {
		msg += "\n⚠️ У вас есть активная бронь\n"
	}

	buttons = append(buttons, moduleRefButton(did, domain.ModuleLaundry, svc)...)
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleLaundrySlots(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateLaundryViewQueue
	fsmMgr.Set(ctx, userCtx)

	did, _ := uuid.Parse(userCtx.DormitoryID)
	uid, _ := uuid.Parse(userCtx.UserUUID)
	today := timeutil.Today()

	// Check if advance booking is available for this user.
	maxDays := 0
	settings, _ := svc.LaundryService.GetLaundrySettings(did)
	if settings != nil && settings.AdvanceBookingEnabled {
		// Check RBAC: does user have advance_booking permission?
		hasAdvance := false
		if userCtx.EmployeeUUID != "" {
			eid, _ := uuid.Parse(userCtx.EmployeeUUID)
			hasAdvance = svc.Guards.CheckPermission(did, eid, "laundry", "read")
		} else if userCtx.UserUUID != "" {
			hasAdvance = svc.Guards.CheckPermissionForUser(did, uid, "laundry", "read")
		}
		if hasAdvance {
			maxDays = settings.AdvanceBookingMaxDays
		}
	}

	if maxDays > 0 {
		// Show date picker.
		layout := "2006-01-02"
		base, _ := time.Parse(layout, today)
		btns := [][]keyboards.Button{}
		for i := 0; i <= maxDays; i++ {
			d := base.AddDate(0, 0, i).Format(layout)
			label := timeutil.FormatEU(d)
			if i == 0 {
				label = "📅 Сегодня (" + label + ")"
			}
			btns = append(btns, []keyboards.Button{{
				Text:    label,
				Payload: fmt.Sprintf("laundry:slots:%s", d),
			}})
		}
		btns = append(btns, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
		return types.ResponseWithButtons("🕐 Выберите дату:", btns), nil
	}

	// Default: today only.
	return showLaundrySlotsForDate(ctx, userCtx, svc, did, today)
}

func showLaundrySlotsForDate(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, did uuid.UUID, date string) (*types.MessageResponse, error) {
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	if len(machines) == 0 {
		return types.ResponseWithButtons("👕 Нет стиральных машин.", keyboards.GetBackKeyboard()), nil
	}

	floors, _ := svc.ResidentService.GetFloorsByDormitory(did)
	floorMap := make(map[uuid.UUID]int)
	for _, f := range floors {
		floorMap[f.ID] = f.FloorNumber
	}

	allBookings, _ := svc.LaundryService.GetDormitoryBookings(did, date)

	bookingByMachine := make(map[uuid.UUID][]domain.LaundryBooking)
	for _, b := range allBookings {
		if b.Status == "active" && b.SlotDate == date {
			bookingByMachine[b.MachineID] = append(bookingByMachine[b.MachineID], b)
		}
	}

	buttons := [][]keyboards.Button{}

	dateLabel := timeutil.FormatEU(date)
	if date == timeutil.Today() {
		dateLabel = "сегодня"
	}
	msg := fmt.Sprintf("🕐 Стиральные машины на %s:\n\n", dateLabel)

	for _, m := range machines {
		status := "✅ свободна"
		if !m.IsActive {
			status = "⛔ выключена"
		} else if bList, ok := bookingByMachine[m.ID]; ok && len(bList) > 0 {
			status = fmt.Sprintf("🔴 занята (%s–%s)", bList[0].SlotStart, bList[0].SlotEnd)
		}
		msg += fmt.Sprintf("• №%s (этаж %d, %d мин) — %s\n", m.MachineNumber, floorMap[m.FloorID], m.DurationMinutes, status)

		buttons = append(buttons, []keyboards.Button{{
			Text:    fmt.Sprintf("Машина %s (этаж %d)", m.MachineNumber, floorMap[m.FloorID]),
			Payload: fmt.Sprintf("laundry:machine:%s:%s", m.ID.String(), date),
		}})
	}

	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleLaundrySlotsDate(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	date := timeutil.Today()

	did, _ := uuid.Parse(userCtx.DormitoryID)
	machines, _ := svc.LaundryService.GetWashingMachines(did)

	buttons := [][]keyboards.Button{}
	for _, m := range machines {
		bookings, _ := svc.LaundryService.GetDormitoryBookings(did, date)
		status := "✅ свободна"
		for _, b := range bookings {
			if b.MachineID == m.ID && b.Status == "active" && b.SlotDate == date {
				status = fmt.Sprintf("🔴 занята (%s–%s)", b.SlotStart, b.SlotEnd)
				break
			}
		}
		if !m.IsActive {
			status = "⛔ выключена"
		}
		buttons = append(buttons, []keyboards.Button{{
			Text:    fmt.Sprintf("Машина %s — %s", m.MachineNumber, status),
			Payload: fmt.Sprintf("laundry:machine:%s:%s", m.ID.String(), date),
		}})
	}

	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(fmt.Sprintf("🕐 Машины на %s:", date), buttons), nil
}

func handleLaundryMachineSlots(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	parts := strings.Split(strings.TrimPrefix(payload, "laundry:machine:"), ":")
	machineID := parts[0]
	date := parts[1]
	userCtx.ContextData["laundry_date"] = date

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

	var blockedStart, blockedEnd string
	blockedStart = target.BlockedSlotStart
	blockedEnd = target.BlockedSlotEnd

	endTimeStr := "23:00"
	if settings != nil && settings.BookingEndTime != "" {
		endTimeStr = settings.BookingEndTime
	}

	var defaultStartHour, defaultStartMin int
	fmt.Sscanf(startTime, "%d:%d", &defaultStartHour, &defaultStartMin)
	var endHour, endMin int
	fmt.Sscanf(endTimeStr, "%d:%d", &endHour, &endMin)
	slotStart := time.Date(0, 1, 1, defaultStartHour, defaultStartMin, 0, 0, time.UTC)
	slotEndTime := time.Date(0, 1, 1, endHour, endMin, 0, 0, time.UTC)

	allBookings, _ := svc.LaundryService.GetDormitoryBookings(did, date)

	var items []keyboards.PageableButton
	now := timeutil.Now()

	for slotStart.Before(slotEndTime) {
		slotEnd := slotStart.Add(time.Duration(duration) * time.Minute)
		if slotEnd.After(slotEndTime) {
			break
		}

		startStr := slotStart.Format("15:04")
		endStr := slotEnd.Format("15:04")

		isPast := date == now.Format("2006-01-02") && startStr <= now.Format("15:04")

		isBlocked := false
		if blockedStart != "" && startStr < blockedEnd && endStr > blockedStart {
			isBlocked = true
		}

		conflict := false
		for _, b := range allBookings {
			if b.MachineID == mid && b.SlotStart < endStr && b.SlotEnd > startStr && b.Status == "active" {
				conflict = true
				break
			}
		}

		if isBlocked {
			items = append(items, keyboards.PageableButton{
				Text:    fmt.Sprintf("👔 %s–%s (кастелянша)", startStr, endStr),
				Payload: "none",
			})
		} else if !conflict && !isPast {
			items = append(items, keyboards.PageableButton{
				Text:    fmt.Sprintf("✅ %s–%s", startStr, endStr),
				Payload: keyboards.PayloadLaundryBookSlot(machineID, startStr),
			})
		} else if !isPast {
			items = append(items, keyboards.PageableButton{
				Text:    fmt.Sprintf("🔴 %s–%s (занято)", startStr, endStr),
				Payload: "none",
			})
		}

		slotStart = slotEnd
	}

	prefix := fmt.Sprintf("laundry:machine:%s:%s", machineID, date)
	buttons := keyboards.BuildPaginatedKeyboard(items, prefix, page, 8)

	return types.ResponseWithButtons(fmt.Sprintf("🕐 Машина %s, %s:", target.MachineNumber, date), buttons), nil
}

func handleLaundryMyBookings(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	uid, _ := uuid.Parse(userCtx.UserUUID)
	bookings, err := svc.LaundryService.GetUserAllBookings(uid)

	if err != nil || len(bookings) == 0 {
		return types.ResponseWithButtons("👕 У вас нет активных броней.", keyboards.GetBackKeyboard()), nil
	}

	today := timeutil.Today()
	buttons := [][]keyboards.Button{}
	msg := "📅 Мои брони:\n\n"
	hasActive := false

	for _, b := range bookings {
		if b.Status != "active" || b.SlotDate < today {
			continue
		}
		hasActive = true
		msg += fmt.Sprintf("• %s: %s–%s (статус: %s)\n", timeutil.FormatEU(b.SlotDate), b.SlotStart, b.SlotEnd, b.Status)
		if b.Status == "active" {
			buttons = append(buttons, []keyboards.Button{{
				Text:    fmt.Sprintf("❌ Отменить %s %s–%s", timeutil.FormatEU(b.SlotDate), b.SlotStart, b.SlotEnd),
				Payload: fmt.Sprintf("laundry:cancel:%s", b.ID.String()),
			}})
		}
	}

	if !hasActive {
		return types.ResponseWithButtons("👕 У вас нет активных броней.", keyboards.GetBackKeyboard()), nil
	}

	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleLaundryViewQueue(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	return handleMenuLaundry(ctx, userCtx, svc, fsmMgr, page)
}

func handleLaundryBook(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	payload = strings.TrimPrefix(payload, "laundry:book:")
	idx := strings.Index(payload, ":")
	if idx < 0 {
		return types.TextResponse("❌ Ошибка выбора слота."), nil
	}
	machineID, slotStart := payload[:idx], payload[idx+1:]

	mid, _ := uuid.Parse(machineID)
	uid, _ := uuid.Parse(userCtx.UserUUID)
	did, _ := uuid.Parse(userCtx.DormitoryID)
	date := timeutil.Today()
	if d, ok := userCtx.ContextData["laundry_date"].(string); ok && d != "" {
		date = d
	}
	userCtx.ContextData["laundry_date"] = nil

	machines, _ := svc.LaundryService.GetWashingMachines(did)
	var machineNumber string
	var duration int
	for _, m := range machines {
		if m.ID == mid {
			machineNumber = m.MachineNumber
			duration = m.DurationMinutes
			break
		}
	}
	if duration == 0 {
		duration = 60
	}

	slotEnd := calculateSlotEnd(slotStart, duration)

	bt, _ := userCtx.ContextData["laundry_booker_type"].(string)
	if bt == "" {
		bt = "resident"
	}
	booking, err := svc.LaundryService.CreateBooking(uid, bt, mid, date, slotStart, slotEnd)
	if err != nil {
		log.Printf("Error creating laundry booking: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	svc.NotificationService.ScheduleBookingReminders(ctx, machineNumber, booking.SlotDate, booking.SlotStart, booking.SlotEnd, userCtx.UserID)

	userCtx.PushState()
	userCtx.State = fsm.StateLaundry
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	msg := fmt.Sprintf(templates.MsgLaundryBookSuccess, booking.MachineNumber, booking.SlotDate, booking.SlotStart, booking.SlotEnd)
	return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil
}

func handleLaundryCancel(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	bookingID := strings.TrimPrefix(payload, "laundry:cancel:")

	msg := "❌ Вы уверены, что хотите отменить бронь?"

	return types.ResponseWithButtons(msg, keyboards.GetConfirmationKeyboard(
		"laundry:cancel_confirm:"+bookingID,
		keyboards.PayloadNavBack,
	)), nil
}

func handleLaundryCancelConfirm(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	bookingID := strings.TrimPrefix(payload, "laundry:cancel_confirm:")

	bid, _ := uuid.Parse(bookingID)
	uid, _ := uuid.Parse(userCtx.UserUUID)

	svc.NotificationService.CancelRemindersForBooking(ctx, bid.String())

	if err := svc.LaundryService.CancelBooking(bid, uid); err != nil {
		log.Printf("Error cancelling laundry booking: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateLaundry
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons(templates.MsgLaundryCancelSuccess, keyboards.GetBackKeyboard()), nil
}

func handleMenuRoom(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if page == 0 {
		userCtx.PushState()
		userCtx.State = fsm.StateRoom
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	rid, _ := uuid.Parse(userCtx.RoomID)
	did, _ := uuid.Parse(userCtx.DormitoryID)
	residentUUID, _ := uuid.Parse(userCtx.ResidentUUID)

	resident, _ := svc.RoomService.GetResident(residentUUID)
	items, err := svc.RoomService.GetRoomItems(rid)

	msg := fmt.Sprintf("🚪 Комната %s\n\n", userCtx.RoomNumber)

	if resident != nil && resident.ContractNumber != nil {
		msg += "📋 Договор койко-места\n"
		msg += fmt.Sprintf("Номер: %s\n", *resident.ContractNumber)
		if resident.ContractStartDate != nil && resident.ContractEndDate != nil {
				msg += fmt.Sprintf("Период: %s – %s\n", timeutil.FormatEU(*resident.ContractStartDate), timeutil.FormatEU(*resident.ContractEndDate))
		}
	} else {
		msg += "📋 Договор не найден в БД университета\n"
	}
	msg += "\n"

	if err != nil || len(items) == 0 {
		controls, _ := svc.ControlService.GetRoomControls(rid)
		if ctrlText := formatControlsForResident(controls); ctrlText != "" {
			msg += ctrlText
		}
		buttons := moduleRefButton(did, domain.ModuleRoom, svc)
		buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
		return types.ResponseWithButtons(msg, buttons), nil
	}

	if err == nil && len(items) > 0 {
		msg += fmt.Sprintf("🪑 Мат. ответственность: %d предметов\n", len(items))
	}

	controls, _ := svc.ControlService.GetRoomControls(rid)
	if ctrlText := formatControlsForResident(controls); ctrlText != "" {
		msg += ctrlText + "\n"
	}

	buttons := moduleRefButton(did, domain.ModuleRoom, svc)
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
	return types.ResponseWithButtons(msg, buttons), nil
}

func handleRoomViewItems(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	return handleMenuRoom(ctx, userCtx, svc, fsmMgr, page)
}

func handleMenuInfo(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if page == 0 {
		userCtx.PushState()
		userCtx.State = fsm.StateInfo
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	materials, err := svc.ReferenceService.GetByModule(did, domain.ModuleReferences)
	if err != nil || len(materials) == 0 {
		return types.ResponseWithButtons("📖 Справочных материалов пока нет.", keyboards.GetBackKeyboard()), nil
	}

	items := make([]keyboards.PageableButton, 0, len(materials))
	for _, m := range materials {
		items = append(items, keyboards.PageableButton{
			Text:    fmt.Sprintf("%s [%s]", m.Name, m.Category),
			Payload: keyboards.PayloadInfoViewItem(m.ID.String()),
		})
	}

	msg := "📖 Справки:\n\n"
	for _, m := range materials {
		msg += fmt.Sprintf("• %s [%s]\n", m.Name, m.Category)
	}

	return types.ResponseWithButtons(msg, keyboards.BuildPaginatedKeyboard(items, "info:list", page, keyboards.DefaultPageSize)), nil
}

func handleInfoViewList(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	return handleMenuInfo(ctx, userCtx, svc, fsmMgr, page)
}

func handleInfoViewItem(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	materialID := strings.TrimPrefix(payload, "info:view:")
	mid, _ := uuid.Parse(materialID)
	material, err := svc.ReferenceService.GetByID(mid)
	if err != nil {
		return types.TextResponse("📖 Материал не найден."), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateInfoViewItem
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	msg := fmt.Sprintf(templates.MsgInfoItem, material.Name, material.Description)
	buttons := [][]keyboards.Button{
		{{Text: "⬅️ Назад к списку", Payload: keyboards.PayloadNavBack}},
	}
	return types.ResponseWithButtons(msg, buttons), nil
}

func handleMenuChatLinks(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateChatLinks
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	did, _ := uuid.Parse(userCtx.DormitoryID)
	links, err := svc.ChatLinkService.GetForResident(did, uuid.Nil, nil)

	msg := "💬 Чаты общежития\n\n"
	if err != nil || len(links) == 0 {
		msg += templates.MsgChatLinksEmpty
		buttons := moduleRefButton(did, domain.ModuleChatLinks, svc)
		buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
		return types.ResponseWithButtons(msg, buttons), nil
	}

	var linksText string
	for _, l := range links {
		linksText += templates.FormatChatLink(l.Title, l.URL, l.Platform) + "\n"
	}
	msg += linksText

	buttons := moduleRefButton(did, domain.ModuleChatLinks, svc)
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})
	return types.ResponseWithButtons(msg, buttons), nil
}

func handleMenuCleaning(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	return showResidentCleaning(ctx, userCtx, svc, fsmMgr, nil)
}

func showResidentCleaning(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, monthOffset *int) (*types.MessageResponse, error) {
	userCtx.PushState()
	userCtx.State = fsm.StateCleaning
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	rid, _ := uuid.Parse(userCtx.RoomID)
	did, _ := uuid.Parse(userCtx.DormitoryID)
	now := timeutil.Now()
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

	duties, _ := svc.CleaningService.GetResidentSchedule(rid, m, y)

	penalties, _ := svc.CleaningService.GetActivePenaltiesByRoom(rid)
	penaltyCounts := make(map[uuid.UUID]int)
	for _, p := range penalties {
		penaltyCounts[p.ID] = p.Count - p.CompletedCount
	}
	penaltyProgress := make(map[uuid.UUID]int)
	for _, p := range penalties {
		penaltyProgress[p.ID] = 0
	}

	monthNames := []string{"", "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"}

	typeMap := map[string]string{"regular": "🧹 Очередное", "penalty": "⚠️ Штрафное", "kitchen": "🍳 Кухня", "garbage": "🗑️ Мусор"}

	msg := fmt.Sprintf("🧹 Дежурства — %s %d\n\n", monthNames[m], y)

	if len(duties) == 0 {
		msg += "Дежурств на этот месяц нет."
	} else {
		for _, d := range duties {
			statusStr := ""
			switch d.Status {
			case "completed":
				statusStr = " ✅"
			case "missed":
				statusStr = " ❌"
			}
			line := fmt.Sprintf("• %s — %s", timeutil.FormatEU(d.DutyDate), typeMap[d.DutyType])
			if d.DutyType == "penalty" && d.PenaltyCleaningID != nil {
				penaltyProgress[*d.PenaltyCleaningID]++
				total := penaltyCounts[*d.PenaltyCleaningID]
				if total > 0 {
					line += fmt.Sprintf(" (%d из %d)", penaltyProgress[*d.PenaltyCleaningID], total)
				}
			}
			line += statusStr + "\n"
			msg += line
		}
	}

	buttons := [][]keyboards.Button{}

	for _, d := range duties {
		if d.Status == "scheduled" {
			buttons = append(buttons, []keyboards.Button{{
				Text:    fmt.Sprintf("✅ Отметить %s", timeutil.FormatEU(d.DutyDate)),
				Payload: fmt.Sprintf("cleaning:resident_complete:%s", d.ID.String()),
			}})
		}
	}

	navRow := []keyboards.Button{
		{Text: "◀", Payload: "cleaning:resident_month:-1"},
		{Text: fmt.Sprintf("%s %d", monthNames[m], y), Payload: "cleaning:resident_month:0"},
		{Text: "▶", Payload: "cleaning:resident_month:1"},
	}
	buttons = append(buttons, navRow)

	buttons = append(buttons, moduleRefButton(did, domain.ModuleCleaning, svc)...)
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func handleCleaningResidentComplete(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	dutyID := strings.TrimPrefix(payload, "cleaning:resident_complete:")
	did, _ := uuid.Parse(dutyID)
	residentUUID, _ := uuid.Parse(userCtx.ResidentUUID)

	if err := svc.CleaningService.MarkDutyCompleted(did, residentUUID); err != nil {
		log.Printf("Error marking duty completed: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}

	return types.ResponseWithButtons("✅ Дежурство отмечено как выполненное!", keyboards.GetBackKeyboard()), nil
}

func handleCleaningResidentMonth(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	offsetStr := strings.TrimPrefix(payload, "cleaning:resident_month:")
	offset, _ := strconv.Atoi(offsetStr)
	return showResidentCleaning(ctx, userCtx, svc, fsmMgr, &offset)
}

func handleCleaningView(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	return handleMenuCleaning(ctx, userCtx, svc, fsmMgr)
}

func moduleRefButton(did uuid.UUID, moduleKey string, svc *service.DBService) [][]keyboards.Button {
	refs, _ := svc.ReferenceService.GetByModule(did, moduleKey)
	if len(refs) == 0 {
		return nil
	}
	return [][]keyboards.Button{{{Text: fmt.Sprintf("📖 Справки (%d)", len(refs)), Payload: "module_refs:" + moduleKey}}}
}

func handleModuleRefs(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	if page == 0 {
		moduleKey := strings.TrimPrefix(payload, "module_refs:")
		userCtx.PushState()
		userCtx.State = fsm.StateModuleRefs
		userCtx.ContextData = map[string]interface{}{"module_key": moduleKey}
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
	}
	return renderModuleRefs(ctx, userCtx, svc, fsmMgr, page)
}

func handleModuleRefsPaged(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	return renderModuleRefs(ctx, userCtx, svc, fsmMgr, page)
}

func renderModuleRefs(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {

	moduleKey, _ := userCtx.ContextData["module_key"].(string)

	did, _ := uuid.Parse(userCtx.DormitoryID)
	refs, _ := svc.ReferenceService.GetByModule(did, moduleKey)

	if len(refs) == 0 {
		return types.ResponseWithButtons("📖 Нет справок в этом разделе.", keyboards.GetBackKeyboard()), nil
	}

	items := make([]keyboards.PageableButton, 0, len(refs))
	for _, r := range refs {
		items = append(items, keyboards.PageableButton{
			Text:    r.Name,
			Payload: fmt.Sprintf("module_ref_view:%s", r.ID.String()),
		})
	}

	kb := keyboards.BuildPaginatedKeyboard(items, "module_refs:list", page, keyboards.DefaultPageSize)
	return types.ResponseWithButtons("📖 Справки раздела:", kb), nil
}

func handleModuleRefView(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	refID := strings.TrimPrefix(payload, "module_ref_view:")
	id, _ := uuid.Parse(refID)
	material, err := svc.ReferenceService.GetByID(id)
	if err != nil {
		return types.ResponseWithButtons("📖 Справка не найдена.", keyboards.GetBackKeyboard()), nil
	}

	userCtx.PushState()
	userCtx.State = fsm.StateModuleRefView
	fsmMgr.Set(ctx, userCtx)

	msg := fmt.Sprintf("📖 %s\n\nКатегория: %s\n\n%s", material.Name, material.Category, material.Description)
	return types.ResponseWithButtons(msg, keyboards.GetBackKeyboard()), nil
}

func handleNavBack(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	prevState := userCtx.PopState()
	userCtx.State = prevState
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	switch prevState {
	case fsm.StateAuthorized:
		if userCtx.EmployeeUUID != "" {
			kb := newAdminMenuBuilderFromContext(svc, userCtx)
			userCtx.State = fsm.StateRoleVerified
			fsmMgr.Set(ctx, userCtx)
			return types.ResponseWithButtons(
				templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
				kb.build(ctx),
			), nil
		}
		return types.ResponseWithButtons(templates.MsgMainMenu, buildResidentMainMenu(ctx, userCtx, svc)), nil
	case fsm.StateAuthorizedEmployee, fsm.StateRoleVerified:
		kb := newAdminMenuBuilderFromContext(svc, userCtx)
		return types.ResponseWithButtons(
			templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
			kb.build(ctx),
		), nil
	default:
		if userCtx.EmployeeUUID != "" {
			kb := newAdminMenuBuilderFromContext(svc, userCtx)
			return types.ResponseWithButtons(
				templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
				kb.build(ctx),
			), nil
		}
		return types.ResponseWithButtons(templates.MsgMainMenu, buildResidentMainMenu(ctx, userCtx, svc)), nil
	}
}

func handleNavHome(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.EmployeeUUID != "" && userCtx.EmployeeRole != "" {
		userCtx.State = fsm.StateAuthorizedEmployee
	} else {
		userCtx.State = fsm.StateAuthorized
	}
	userCtx.ClearNavigationStack()
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	if userCtx.State == fsm.StateAuthorized {
		return types.ResponseWithButtons(templates.MsgMainMenu, buildResidentMainMenu(ctx, userCtx, svc)), nil
	}
	kb := newAdminMenuBuilderFromContext(svc, userCtx)
	return types.ResponseWithButtons(
		templates.FormatAuthorizedEmployee(userCtx.FirstName, kb.RoleDisplay()),
		kb.build(ctx),
	), nil
}

func calculateSlotEnd(startTime string, durationMin int) string {
	var h, m int
	fmt.Sscanf(startTime, "%d:%d", &h, &m)
	m += durationMin
	h += m / 60
	m = m % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func handleResidentRoleSelect(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	uid, _ := uuid.Parse(userCtx.UserUUID)
	did, _ := uuid.Parse(userCtx.DormitoryID)

	roles, err := svc.GetUserRolesForDormitory(uid, did)
	if err != nil || len(roles) == 0 {
		return types.ResponseWithButtons("🔑 У вас нет назначенных ролей.", keyboards.GetBackKeyboard()), nil
	}

	if len(roles) == 1 {
		r := roles[0]
		userCtx.PushState()
		userCtx.State = fsm.StateResidentRoleVerified
		userCtx.EmployeeRole = string(r.Role)
		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}
		rolenames := buildRoleDisplayLookup(svc)
		roleDisplay := service.MapRoleDisplayWithLookup(string(r.Role), rolenames)
		msg := fmt.Sprintf("🔑 Вы вошли как «%s».", roleDisplay)
		return types.ResponseWithButtons(msg, newAdminMenuBuilderFromContext(svc, userCtx).build(ctx)), nil
	}

	rolenames := buildRoleDisplayLookup(svc)
	buttons := [][]keyboards.Button{}
	for _, r := range roles {
		roleDisplay := service.MapRoleDisplayWithLookup(string(r.Role), rolenames)
		buttons = append(buttons, []keyboards.Button{{
			Text:    roleDisplay,
			Payload: fmt.Sprintf("resident_role:confirm:%s", string(r.Role)),
		}})
	}
	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons("🔑 Выберите роль:", buttons), nil
}

func handleResidentRoleConfirm(ctx context.Context, userCtx *fsm.UserContext, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	roleName := strings.TrimPrefix(payload, "resident_role:confirm:")

	userCtx.PushState()
	userCtx.State = fsm.StateResidentRoleVerified
	userCtx.EmployeeRole = roleName
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	rolenames := buildRoleDisplayLookup(svc)
	roleDisplay := service.MapRoleDisplayWithLookup(roleName, rolenames)
	msg := fmt.Sprintf("🔑 Вы вошли как «%s».", roleDisplay)
	return types.ResponseWithButtons(msg, newAdminMenuBuilderFromContext(svc, userCtx).build(ctx)), nil
}

func handleResidentRoleExit(ctx context.Context, userCtx *fsm.UserContext, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	userCtx.EmployeeRole = ""
	userCtx.State = userCtx.PopState()
	if userCtx.State == "" {
		userCtx.State = fsm.StateAuthorized
		userCtx.ClearNavigationStack()
	}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}

	return types.ResponseWithButtons(templates.MsgMainMenu, buildResidentMainMenu(ctx, userCtx, svc)), nil
}
