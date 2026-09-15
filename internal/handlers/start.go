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

func HandleStartCommand(ctx context.Context, maxUserID int64, fsmMgr *fsm.FSMManager, svc *service.DBService) (*types.MessageResponse, error) {
	userCtx, err := fsmMgr.Get(ctx, maxUserID)
	if err != nil {
		log.Printf("Error getting user context: %v", err)
		return nil, err
	}

	var messageText string
	var buttons [][]keyboards.Button

	switch userCtx.State {
	case fsm.StateUninitialized:
		messageText = templates.MsgWelcome
		userCtx.State = fsm.StateAwaitingContact
		buttons = keyboards.GetContactKeyboard()
	case fsm.StateAuthorized:
		userName := userCtx.FirstName
		if userCtx.LastName != "" {
			userName = userCtx.FirstName + " " + userCtx.LastName
		}
		messageText = fmt.Sprintf("👋 С возвращением, %s!\n\n%s", userName, templates.MsgMainMenu)
		buttons = buildResidentMainMenu(ctx, userCtx, svc)
	case fsm.StateAuthorizedEmployee:
		userName := userCtx.FirstName
		if userCtx.LastName != "" {
			userName = userCtx.FirstName + " " + userCtx.LastName
		}
		messageText = fmt.Sprintf("👔 С возвращением, %s!", userName)
		buttons = append(buttons, []keyboards.Button{{Text: "🏢 Выбрать общежитие", Payload: "dormitory:list"}})
	default:
		messageText = templates.MsgWelcome
		userCtx.State = fsm.StateAwaitingContact
		buttons = keyboards.GetContactKeyboard()
	}

	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		log.Printf("Error saving user context: %v", err)
		return nil, err
	}

	return types.ResponseWithButtons(messageText, buttons), nil
}

func HandleUserAdded(ctx context.Context, maxUserID int64, fsmMgr *fsm.FSMManager) error {
	userCtx := &fsm.UserContext{
		State:        fsm.StateUninitialized,
		UserID:       maxUserID,
		IsAuthorized: false,
	}

	if err := fsmMgr.Set(ctx, userCtx); err != nil {
		log.Printf("Error initializing user %d: %v", maxUserID, err)
		return err
	}

	log.Printf("User %d added to bot", maxUserID)
	return nil
}

func HandleUserRemoved(ctx context.Context, maxUserID int64, fsmMgr *fsm.FSMManager) error {
	if err := fsmMgr.Delete(ctx, maxUserID); err != nil {
		log.Printf("Error deleting user context for %d: %v", maxUserID, err)
		return err
	}

	log.Printf("User %d removed from bot", maxUserID)
	return nil
}
