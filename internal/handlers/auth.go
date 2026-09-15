package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/keyboards"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/templates"
	"github.com/dormitory-bot/internal/types"
)

func HandleContactMessage(ctx context.Context, userCtx *fsm.UserContext, phone string, svc *service.DBService, fsmMgr *fsm.FSMManager) (*types.MessageResponse, error) {
	log.Printf("User %d: Processing contact with phone: %s", userCtx.UserID, phone)

	if phone == "" {
		return types.TextResponse("❌ Не удалось распознать номер телефона. Попробуй ещё раз."), nil
	}

	verifyResult, err := svc.VerifyUser(phone, userCtx.UserID, "max")
	if err != nil {
		log.Printf("Error verifying user: %v", err)
		return types.TextResponse(templates.MsgGlobalError), nil
	}

	if !verifyResult.Found {
		return types.TextResponse(templates.MsgVerificationFailed), nil
	}

	userCtx.Phone = phone
	userCtx.FirstName = verifyResult.FirstName
	userCtx.LastName = verifyResult.LastName
	userCtx.MiddleName = verifyResult.MiddleName
	userCtx.EISVerified = verifyResult.EISVerified
	userCtx.EISPersonID = verifyResult.EISPersonID
	userCtx.PersonType = verifyResult.PersonType
	userCtx.UserUUID = verifyResult.UserID

	if verifyResult.Resident != nil {
		userCtx.State = fsm.StateResidentConfirmData
		userCtx.DormitoryID = verifyResult.Resident.DormitoryID
		userCtx.DormitoryName = verifyResult.Resident.DormitoryName
		userCtx.RoomID = verifyResult.Resident.RoomID
		userCtx.RoomNumber = verifyResult.Resident.RoomNumber
		userCtx.FloorNumber = verifyResult.Resident.FloorNumber
		userCtx.ResidentUUID = verifyResult.Resident.ResidentID
		userCtx.IsAuthorized = true

		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}

		msg := templates.FormatResidentData(userCtx.DormitoryName, userCtx.RoomNumber, userCtx.FloorNumber)
		return types.ResponseWithButtons(msg, keyboards.GetConfirmationKeyboard(keyboards.PayloadResidentConfirm, keyboards.PayloadResidentReject)), nil
	}

	if verifyResult.Employee != nil {
		userCtx.State = fsm.StateRoleVerified
		userCtx.DormitoryID = verifyResult.DormitoryID
		userCtx.EmployeeUUID = verifyResult.Employee.EmployeeID
		userCtx.EmployeeRole = verifyResult.Employee.Role
		userCtx.IsAuthorized = true

		rolenames := buildRoleDisplayLookup(svc)
		roleDisplay := service.MapRoleDisplayWithLookup(verifyResult.Employee.Role, rolenames)

		if err := fsmMgr.Set(ctx, userCtx); err != nil {
			return nil, err
		}

		if verifyResult.DormitoryID != "" {
			dorm, _ := svc.GetDormitoryName(verifyResult.DormitoryID)
			msg := fmt.Sprintf("✅ %s, ваша должность — %s.\nОбщежитие: %s", userCtx.FirstName, roleDisplay, dorm)
			return types.ResponseWithButtons(msg, newAdminMenuBuilderFromContext(svc, userCtx).build(ctx)), nil
		}

		dorms, _ := svc.GetEmployeeDormitories(verifyResult.Employee.EmployeeID)
		if len(dorms) > 0 {
			if len(dorms) == 1 {
				userCtx.DormitoryID = dorms[0].DormitoryID
				fsmMgr.Set(ctx, userCtx)
				dorm, _ := svc.GetDormitoryName(dorms[0].DormitoryID)
				msg := fmt.Sprintf("✅ %s, ваша должность — %s.\nОбщежитие: %s", userCtx.FirstName, roleDisplay, dorm)
				return types.ResponseWithButtons(msg, newAdminMenuBuilderFromContext(svc, userCtx).build(ctx)), nil
			}
			var buttons [][]keyboards.Button
			for _, d := range dorms {
				buttons = append(buttons, []keyboards.Button{{Text: d.DormitoryName, Payload: "dormitory:select:" + d.DormitoryID}})
			}
			msg := fmt.Sprintf("✅ %s, ваша должность — %s.\nВыберите общежитие:", userCtx.FirstName, roleDisplay)
			return types.ResponseWithButtons(msg, buttons), nil
		}

		msg := fmt.Sprintf("✅ %s, ваша должность — %s.", userCtx.FirstName, roleDisplay)
		return types.ResponseWithButtons(msg, [][]keyboards.Button{
			{keyboards.Button{Text: "🏢 Выбрать общежитие", Payload: "dormitory:list"}},
		}), nil
	}

	// Ни resident, ни employee не найдены.
	// Не предлагаем выбор роли — пользователь не сможет продолжить.
	log.Printf("User %d: телефон подтверждён, но resident/employee данные не найдены (PersonType=%s)",
		userCtx.UserID, verifyResult.PersonType)
	return types.TextResponse(templates.MsgVerificationFailed), nil
}
