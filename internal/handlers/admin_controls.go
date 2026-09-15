package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/keyboards"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/templates"
	"github.com/dormitory-bot/internal/timeutil"
	"github.com/dormitory-bot/internal/types"
	"github.com/google/uuid"
)

type controlType int

const (
	ctrlSanitary controlType = iota
	ctrlDiscipline
	ctrlCohabitation
)

func handleControlSet(ctx context.Context, userCtx *fsm.UserContext, ctrlType controlType, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "room", "create"); !ok {
		return blocked, nil
	}

	roomID := strings.TrimPrefix(payload, "control:sanitary:set:")
	roomID = strings.TrimPrefix(roomID, "control:discipline:set:")
	roomID = strings.TrimPrefix(roomID, "control:cohabitation:set:")
	rid, _ := uuid.Parse(roomID)

	userCtx.PushState()
	userCtx.State = fsm.StateAwaitingControlData
	userCtx.ContextData = map[string]interface{}{
		"step":      "start_date",
		"ctrl_type": int(ctrlType),
		"room_id":   rid.String(),
	}
	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		return nil, err
	}
	return types.ResponseWithButtons("Введите дату начала (ГГГГ-ММ-ДД):", keyboards.GetBackKeyboard()), nil
}

func HandleAdminControlMessage(ctx context.Context, userCtx *fsm.UserContext, msg string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if userCtx.State != fsm.StateAwaitingControlData {
		return nil, nil
	}

	step, _ := userCtx.ContextData["step"].(string)
	ctrl := controlType(int(userCtx.ContextData["ctrl_type"].(float64)))
	rid, _ := uuid.Parse(userCtx.ContextData["room_id"].(string))

	switch step {
	case "start_date":
		userCtx.ContextData["start_date"] = msg
		userCtx.ContextData["step"] = "end_date"
		fsmMgr.Set(ctx, userCtx)
		return types.ResponseWithButtons("Введите дату окончания (ГГГГ-ММ-ДД):", keyboards.GetBackKeyboard()), nil

	case "end_date":
		userCtx.ContextData["end_date"] = msg

		switch ctrl {
		case ctrlSanitary:
			userCtx.ContextData["step"] = "reason"
			fsmMgr.Set(ctx, userCtx)
			return types.ResponseWithButtons("Введите причину санконтроля:", keyboards.GetBackKeyboard()), nil

		case ctrlDiscipline:
			userCtx.ContextData["step"] = "violation"
			fsmMgr.Set(ctx, userCtx)
			return types.ResponseWithButtons("Введите описание нарушения:", keyboards.GetBackKeyboard()), nil

		case ctrlCohabitation:
			userCtx.ContextData["step"] = "reason"
			fsmMgr.Set(ctx, userCtx)
			return types.ResponseWithButtons("Введите причину контроля проживания:", keyboards.GetBackKeyboard()), nil
		}

	case "reason":
		userCtx.ContextData["reason"] = msg
		eid, _ := uuid.Parse(userCtx.EmployeeUUID)
		startDate := userCtx.ContextData["start_date"].(string)
		endDate := userCtx.ContextData["end_date"].(string)

		switch ctrl {
		case ctrlSanitary:
			_, err := svc.ControlService.SetSanitaryControl(rid, startDate, endDate, msg, eid)
			if err != nil {
				log.Printf("Error setting sanitary control: %v", err)
				return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
			}
		case ctrlCohabitation:
			_, err := svc.ControlService.SetCohabitationControl(rid, startDate, endDate, msg)
			if err != nil {
				log.Printf("Error setting cohabitation control: %v", err)
				return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
			}
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateRoleVerified
		fsmMgr.Set(ctx, userCtx)
		return viewRoom(ctx, userCtx, rid, svc, fsmMgr, 0)

	case "violation":
		userCtx.ContextData["violation"] = msg
		eid, _ := uuid.Parse(userCtx.EmployeeUUID)
		startDate := userCtx.ContextData["start_date"].(string)
		endDate := userCtx.ContextData["end_date"].(string)

		_, err := svc.ControlService.SetDisciplineControl(rid, nil, startDate, endDate, msg, eid)
		if err != nil {
			log.Printf("Error setting discipline control: %v", err)
			return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
		}

		userCtx.ContextData = nil
		userCtx.State = fsm.StateRoleVerified
		fsmMgr.Set(ctx, userCtx)
		return viewRoom(ctx, userCtx, rid, svc, fsmMgr, 0)
	}

	return nil, nil
}

func handleControlRemove(ctx context.Context, userCtx *fsm.UserContext, ctrlType controlType, payload string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	if blocked, ok := guardPermission(userCtx, svc, "room", "delete"); !ok {
		return blocked, nil
	}

	controlID := strings.TrimPrefix(payload, "sanitary_remove:")
	controlID = strings.TrimPrefix(controlID, "discipline_remove:")
	controlID = strings.TrimPrefix(controlID, "cohabitation_remove:")
	cid, _ := uuid.Parse(controlID)

	roomID, _ := uuid.Parse(userCtx.ContextData["room_id"].(string))

	var err error
	switch ctrlType {
	case ctrlSanitary:
		err = svc.ControlService.CompleteSanitaryControl(cid, "Завершено досрочно")
	case ctrlDiscipline:
		err = svc.ControlService.CompleteDisciplineControl(cid, "Завершено досрочно")
	case ctrlCohabitation:
		err = svc.ControlService.CompleteCohabitationControl(cid, "Завершено досрочно")
	}
	if err != nil {
		log.Printf("Error completing control: %v", err)
		return types.ResponseWithButtons(templates.MsgGenericError, keyboards.GetBackKeyboard()), nil
	}
	return viewRoom(ctx, userCtx, roomID, svc, fsmMgr, 0)
}

func viewRoom(ctx context.Context, userCtx *fsm.UserContext, rid uuid.UUID, svc *service.DBService, fsmMgr *fsm.FSMManager, page int) (*types.MessageResponse, error) {
	room, err := svc.RoomService.GetRoomInfo(rid)
	if err != nil {
		return types.TextResponse("🚪 Комната не найдена."), nil
	}

	residents, _ := svc.RoomService.GetRoomResidents(rid)
	controls, _ := svc.ControlService.GetRoomControls(rid)
	materials, _ := svc.RoomService.GetRoomItems(rid)

	userCtx.ContextData = map[string]interface{}{"room_id": rid.String()}
	userCtx.State = fsm.StateRoomManagement
	fsmMgr.Set(ctx, userCtx)

	msg := fmt.Sprintf("🚪 Комната %s\n", room.RoomNumber)
	msg += fmt.Sprintf("Вместимость: %d | Проживает: %d\n", room.Capacity, len(residents))

	if len(materials) > 0 {
		msg += fmt.Sprintf("Мат. ответственность: %d предметов\n", len(materials))
	}

	buttons := [][]keyboards.Button{}

	if len(controls.Sanitary) > 0 {
		for _, c := range controls.Sanitary {
			if c.Status == "active" {
				msg += fmt.Sprintf("\n🧹 Санконтроль: %s – %s", timeutil.FormatEU(c.StartDate), timeutil.FormatEU(c.EndDate))
				if c.Reason != "" {
					msg += fmt.Sprintf(" (%s)", truncate(c.Reason, 30))
				}
				if hasPermission(userCtx, svc, "room", "delete") {
					buttons = append(buttons, []keyboards.Button{{
						Text:    "🧹 Снять с санконтроля",
						Payload: fmt.Sprintf("sanitary_remove:%s", c.ID.String()),
					}})
				}
			}
		}
	}
	if len(controls.Discipline) > 0 {
		for _, c := range controls.Discipline {
			if c.Status == "active" {
				msg += fmt.Sprintf("\n⚠️ Дисц. контроль: %s – %s", timeutil.FormatEU(c.StartDate), timeutil.FormatEU(c.EndDate))
				if c.Violation != "" {
					msg += fmt.Sprintf(" (%s)", truncate(c.Violation, 30))
				}
				if hasPermission(userCtx, svc, "room", "delete") {
					buttons = append(buttons, []keyboards.Button{{
						Text:    "⚠️ Снять с дисц. контроля",
						Payload: fmt.Sprintf("discipline_remove:%s", c.ID.String()),
					}})
				}
			}
		}
	}
	if len(controls.Cohabitation) > 0 {
		for _, c := range controls.Cohabitation {
			if c.Status == "active" {
				msg += fmt.Sprintf("\n🏠 Контроль проживания: %s – %s", timeutil.FormatEU(c.StartDate), timeutil.FormatEU(c.EndDate))
				if hasPermission(userCtx, svc, "room", "delete") {
					buttons = append(buttons, []keyboards.Button{{
						Text:    "🏠 Снять с контроля проживания",
						Payload: fmt.Sprintf("cohabitation_remove:%s", c.ID.String()),
					}})
				}
			}
		}
	}

	buttons = append(buttons, []keyboards.Button{
		{Text: "🔍 Контроль", Payload: fmt.Sprintf("room:controls:%s", rid.String())},
		{Text: "👤 Жильцы", Payload: fmt.Sprintf("room:residents:%s", rid.String())},
	})

	if hasPermission(userCtx, svc, "cleaning", "create") {
		buttons = append(buttons, []keyboards.Button{
			{Text: "⚠️ Назначить штраф", Payload: "penalty:create"},
			{Text: "🔓 Освободить от дежурств", Payload: "exemption:create"},
		})
	}

	if len(residents) > 0 {
		msg += "\n\n👤 Жильцы:\n"
		for _, r := range residents {
			msg += fmt.Sprintf("  • %s %s\n", r.User.LastName, r.User.FirstName)
		}
	}

	buttons = append(buttons, []keyboards.Button{{Text: "⬅️ Назад", Payload: keyboards.PayloadNavBack}})

	return types.ResponseWithButtons(msg, buttons), nil
}

func formatControlsForResident(controls *service.ControlResult) string {
	var msg string
	for _, c := range controls.Sanitary {
		if c.Status == "active" {
			days := daysRemaining(c.EndDate)
			msg += fmt.Sprintf("\n🧹 Санконтроль: %s – %s", timeutil.FormatEU(c.StartDate), timeutil.FormatEU(c.EndDate))
			if days > 0 {
				msg += fmt.Sprintf(" (осталось %d дн.)", days)
			} else {
				msg += " (завершается сегодня)"
			}
			if c.Reason != "" {
				msg += fmt.Sprintf("\n   Причина: %s", c.Reason)
			}
		}
	}
	for _, c := range controls.Discipline {
		if c.Status == "active" {
			days := daysRemaining(c.EndDate)
			msg += fmt.Sprintf("\n⚠️ Дисцип. контроль: %s – %s", timeutil.FormatEU(c.StartDate), timeutil.FormatEU(c.EndDate))
			if days > 0 {
				msg += fmt.Sprintf(" (осталось %d дн.)", days)
			} else {
				msg += " (завершается сегодня)"
			}
			if c.Violation != "" {
				msg += fmt.Sprintf("\n   Нарушение: %s", c.Violation)
			}
		}
	}
	for _, c := range controls.Cohabitation {
		if c.Status == "active" {
			days := daysRemaining(c.EndDate)
			msg += fmt.Sprintf("\n🏠 Контроль проживания: %s – %s", timeutil.FormatEU(c.StartDate), timeutil.FormatEU(c.EndDate))
			if days > 0 {
				msg += fmt.Sprintf(" (осталось %d дн.)", days)
			} else {
				msg += " (завершается сегодня)"
			}
			if c.Reason != "" {
				msg += fmt.Sprintf("\n   Причина: %s", c.Reason)
			}
		}
	}
	return msg
}

func daysRemaining(endDate string) int {
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return -1
	}
	days := int(end.Sub(time.Now()).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
