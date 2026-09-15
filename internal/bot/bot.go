package bot

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/dormitory-bot/internal/fsm"
	"github.com/dormitory-bot/internal/handlers"
	"github.com/dormitory-bot/internal/keyboards"
	"github.com/dormitory-bot/internal/service"
	"github.com/dormitory-bot/internal/types"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type Bot struct {
	fsmMgr      *fsm.FSMManager
	svc         *service.DBService
	maxAPI      *maxbot.Api
	maxAPIToken string
	maxAPIURL   string
	httpClient  *http.Client
	userMu      sync.Map
}

func NewBot(fsmMgr *fsm.FSMManager, svc *service.DBService, maxAPI *maxbot.Api, token, apiURL string) *Bot {
	return &Bot{
		fsmMgr:      fsmMgr,
		svc:         svc,
		maxAPI:      maxAPI,
		maxAPIToken: token,
		maxAPIURL:   apiURL,
		httpClient:  &http.Client{},
	}
}

func (b *Bot) HandleCallback(ctx context.Context, maxUserID int64, payload string, callbackID string) error {
	userCtx, err := b.fsmMgr.Get(ctx, maxUserID)
	if err != nil {
		log.Printf("Error getting user context: %v", err)
		return err
	}

	var response *types.MessageResponse
	switch userCtx.State {
	case fsm.StateAuthorizedEmployee, fsm.StateSelectingDormitory, fsm.StateRoleVerified,
		fsm.StateModulesManagement, fsm.StateLaundryManagement, fsm.StateRoomManagement,
		fsm.StateRoomSearch, fsm.StateCleaningManagement,
		fsm.StateReferenceManagement, fsm.StateChatLinkManagement,
		fsm.StateAwaitingReferenceData, fsm.StateAwaitingChatLinkData,
		fsm.StateAwaitingControlData, fsm.StateAwaitingMachineData,
		fsm.StateAwaitingLaundrySettings, fsm.StateAwaitingAdminBooking,
		fsm.StateAwaitingPenaltyData, fsm.StateAwaitingExemptionData,
		fsm.StateStaffManagement, fsm.StateStaffView, fsm.StateStaffRoleAssign,
		fsm.StateStaffAdd, fsm.StateAwaitingStaffData,
		fsm.StateRoleManagement, fsm.StateRoleView, fsm.StateRoleCreate,
		fsm.StateRoleEdit, fsm.StateAwaitingRoleData,
		fsm.StateResidentRoleVerified, fsm.StateResidentProfile:
		response, err = handlers.HandleAdminCallback(ctx, userCtx, payload, b.svc, b.fsmMgr)
	default:
		response, err = handlers.HandleCallbackQuery(ctx, userCtx, payload, b.svc, b.fsmMgr)
	}
	if err != nil {
		log.Printf("Error handling callback %s: %v", payload, err)
		return err
	}

	if response != nil && response.Text != "" {
		if callbackID != "" {
			if err := b.AnswerCallback(ctx, callbackID, response.Text, response.Buttons); err == nil {
				return nil
			}
			log.Printf("AnswerCallback failed, falling back to SendOrEdit")
		}
		return b.SendOrEdit(ctx, userCtx, response.Text, response.Buttons)
	}

	return nil
}

func (b *Bot) HandleBotStarted(ctx context.Context, maxUserID int64) error {
	response, err := handlers.HandleStartCommand(ctx, maxUserID, b.fsmMgr, b.svc)
	if err != nil {
		return err
	}

	if response != nil && response.Text != "" {
		return b.SendMessageWithKeyboard(ctx, maxUserID, response.Text, response.Buttons)
	}

	return nil
}

func (b *Bot) HandleUserAdded(ctx context.Context, maxUserID int64) error {
	return handlers.HandleUserAdded(ctx, maxUserID, b.fsmMgr)
}

func (b *Bot) HandleUserRemoved(ctx context.Context, maxUserID int64) error {
	return handlers.HandleUserRemoved(ctx, maxUserID, b.fsmMgr)
}

func (b *Bot) SendTextToUser(ctx context.Context, userID int64, text string) error {
	return b.SendMessageWithKeyboard(ctx, userID, text, nil)
}

func (b *Bot) GetPendingReminders(ctx context.Context) ([]string, error) {
	return b.svc.NotificationService.GetPendingReminders(ctx)
}

func (b *Bot) PollingLoop(ctx context.Context) error {
	log.Println("Bot polling loop started")

	for update := range b.maxAPI.GetUpdates(ctx) {
		go b.processUpdate(ctx, update)
	}

	log.Println("Bot polling loop stopped")
	return nil
}

func (b *Bot) getUserMutex(userID int64) *sync.Mutex {
	actual, _ := b.userMu.LoadOrStore(userID, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (b *Bot) processUpdate(ctx context.Context, update interface{}) {
	var userID int64
	switch upd := update.(type) {
	case *schemes.MessageCreatedUpdate:
		userID = upd.GetUserID()
	case *schemes.MessageCallbackUpdate:
		userID = upd.GetUserID()
	case *schemes.BotStartedUpdate:
		userID = upd.GetUserID()
	case *schemes.UserAddedToChatUpdate:
		userID = upd.GetUserID()
	case *schemes.UserRemovedFromChatUpdate:
		userID = upd.GetUserID()
	default:
		b.processUpdateInternal(ctx, update)
		return
	}

	mu := b.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	b.processUpdateInternal(ctx, update)
}

func (b *Bot) processUpdateInternal(ctx context.Context, update interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in processUpdate: %v", r)
		}
	}()

	switch upd := update.(type) {
	case *schemes.MessageCreatedUpdate:
		userID := upd.GetUserID()

		if upd.GetCommand() == "" && upd.GetText() != "" {
			userCtx, err := b.fsmMgr.Get(ctx, userID)
			if err == nil {
				response, err := handlers.HandleAdminMessage(ctx, userCtx, upd.GetText(), b.svc, b.fsmMgr)
				if err == nil && response != nil {
					b.SendOrEdit(ctx, userCtx, response.Text, response.Buttons)
					return
				}
			}
		}

		if len(upd.Message.Body.Attachments) > 0 {
			log.Printf("User %d: Message has %d attachments", userID, len(upd.Message.Body.Attachments))
			for i, att := range upd.Message.Body.Attachments {
				log.Printf("[Attachment %d] Processing attachment, actual type: %T", i, att)
				var phone string

				if contactAtt, ok := att.(*schemes.ContactAttachment); ok {
					log.Printf("✅ [Attachment %d] Contact attachment detected (native type)", i)
					log.Printf("   ContactAttachment type: %v", reflect.TypeOf(contactAtt))
					log.Printf("   ContactAttachment.Payload type: %v", reflect.TypeOf(contactAtt.Payload))

					if jsonBytes, err := json.Marshal(contactAtt.Payload); err == nil {
						log.Printf("   ContactAttachment.Payload JSON: %s", string(jsonBytes))
						var payloadMap map[string]interface{}
						if err := json.Unmarshal(jsonBytes, &payloadMap); err == nil {
							log.Printf("   Available fields in Payload: %v", payloadMap)

							if phoneVal, ok := payloadMap["phone"].(string); ok && phoneVal != "" {
								phone = cleanPhone(phoneVal)
								log.Printf("   Extracted phone from field 'phone': %s", phone)
							} else if phoneVal, ok := payloadMap["phone_number"].(string); ok && phoneVal != "" {
								phone = cleanPhone(phoneVal)
								log.Printf("   Extracted phone from field 'phone_number': %s", phone)
							} else if vcfVal, ok := payloadMap["vcf_info"].(string); ok && vcfVal != "" {
								log.Printf("   Phone not found, trying vcf_info field (%d chars)", len(vcfVal))
								phone = extractPhoneFromVCF(vcfVal)
								log.Printf("   Extracted phone from vcf_info: %s", phone)
							} else if vcfVal, ok := payloadMap["vcard"].(string); ok && vcfVal != "" {
								log.Printf("   Phone not found, trying vcard field (%d chars)", len(vcfVal))
								phone = extractPhoneFromVCF(vcfVal)
								log.Printf("   Extracted phone from vcard: %s", phone)
							} else if vcfVal, ok := payloadMap["vcf"].(string); ok && vcfVal != "" {
								log.Printf("   Phone not found, trying vcf field (%d chars)", len(vcfVal))
								phone = extractPhoneFromVCF(vcfVal)
								log.Printf("   Extracted phone from vcf: %s", phone)
							} else {
								log.Printf("   Could not find phone in any field. Dumping keys:")
								for k := range payloadMap {
									log.Printf("     Key: %s", k)
								}
							}
						}
					} else {
						log.Printf("   Error marshaling Payload: %v", err)
					}
				} else {
					if jsonBytes, err := json.Marshal(att); err == nil {
						s := string(jsonBytes)
						if len(s) > 500 {
							s = s[:500]
						}
						log.Printf("⚠️ [Attachment %d] Non-ContactAttachment type: %T, JSON: %s", i, att, s)
					}
				}

				if phone != "" {
					userCtx, err := b.fsmMgr.Get(ctx, userID)
					if err != nil {
						log.Printf("Error getting user context: %v", err)
						continue
					}

					response, err := handlers.HandleContactMessage(ctx, userCtx, phone, b.svc, b.fsmMgr)
					if err != nil {
						log.Printf("Error handling contact message: %v", err)
						continue
					}
					if response != nil && response.Text != "" {
						b.SendMessageWithKeyboard(ctx, userID, response.Text, response.Buttons)
					}
					return
				}
				log.Printf("Attachment %d: Failed to extract phone", i)
			}
		}

		if upd.GetCommand() == "/start" {
			response, err := handlers.HandleStartCommand(ctx, userID, b.fsmMgr, b.svc)
			if err != nil {
				log.Printf("Error handling /start: %v", err)
			} else if response != nil && response.Text != "" {
				b.SendMessageWithKeyboard(ctx, userID, response.Text, response.Buttons)
			}
		} else if text := upd.GetText(); text != "" {
			userCtx, err := b.fsmMgr.Get(ctx, userID)
			if err != nil {
				log.Printf("Error getting user context: %v", err)
				return
			}
			response, err := handlers.HandleAdminMessage(ctx, userCtx, text, b.svc, b.fsmMgr)
			if err != nil {
				log.Printf("Error handling admin message: %v", err)
				return
			}
			if response != nil && response.Text != "" {
				b.SendOrEdit(ctx, userCtx, response.Text, response.Buttons)
			}
		}

	case *schemes.MessageCallbackUpdate:
		userID := upd.GetUserID()
		payload := upd.Callback.Payload
		callbackID := upd.Callback.CallbackID
		if err := b.HandleCallback(ctx, userID, payload, callbackID); err != nil {
			log.Printf("Error handling callback: %v", err)
		}

	case *schemes.BotStartedUpdate:
		userID := upd.GetUserID()
		if err := b.HandleBotStarted(ctx, userID); err != nil {
			log.Printf("Error handling bot_started: %v", err)
		}

	case *schemes.UserAddedToChatUpdate:
		userID := upd.GetUserID()
		if err := b.HandleUserAdded(ctx, userID); err != nil {
			log.Printf("Error handling user_added: %v", err)
		}

	case *schemes.UserRemovedFromChatUpdate:
		userID := upd.GetUserID()
		if err := b.HandleUserRemoved(ctx, userID); err != nil {
			log.Printf("Error handling user_removed: %v", err)
		}

	default:
		log.Printf("Unknown update type: %T", update)
	}
}

func (b *Bot) SendMessageWithKeyboard(ctx context.Context, maxUserID int64, text string, rows [][]keyboards.Button) error {
	message := maxbot.NewMessage().
		SetUser(maxUserID).
		SetText(text).
		SetFormat("html")

	if len(rows) > 0 {
		kb := b.maxAPI.Messages.NewKeyboardBuilder()
		for _, row := range rows {
			kbRow := kb.AddRow()
			for _, btn := range row {
				if btn.Type == "contact" {
					kbRow.AddContact(btn.Text)
				} else if btn.Type == "link" {
					kbRow.AddLink(btn.Text, schemes.POSITIVE, btn.Payload)
				} else {
					kbRow.AddCallback(btn.Text, schemes.POSITIVE, btn.Payload)
				}
			}
		}
		message.AddKeyboard(kb)
	}

	err := b.maxAPI.Messages.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending message to user %d: %v", maxUserID, err)
		return err
	}
	return nil
}

func (b *Bot) EditMessageWithKeyboard(ctx context.Context, msgID string, text string, rows [][]keyboards.Button) error {
	message := maxbot.NewMessage().SetText(text).SetFormat("html")

	if len(rows) > 0 {
		kb := b.maxAPI.Messages.NewKeyboardBuilder()
		for _, row := range rows {
			kbRow := kb.AddRow()
			for _, btn := range row {
				if btn.Type == "contact" {
					kbRow.AddContact(btn.Text)
				} else if btn.Type == "link" {
					kbRow.AddLink(btn.Text, schemes.POSITIVE, btn.Payload)
				} else {
					kbRow.AddCallback(btn.Text, schemes.POSITIVE, btn.Payload)
				}
			}
		}
		message.AddKeyboard(kb)
	}

	err := b.maxAPI.Messages.EditMessage(ctx, msgID, message)
	if err != nil {
		log.Printf("Error editing message %s: %v", msgID, err)
		return err
	}
	return nil
}

func (b *Bot) SendOrEdit(ctx context.Context, userCtx *fsm.UserContext, text string, rows [][]keyboards.Button) error {
	if userCtx.LastMessageID != "" {
		if err := b.EditMessageWithKeyboard(ctx, userCtx.LastMessageID, text, rows); err != nil {
			log.Printf("Edit failed, falling back to new message: %v", err)
			return b.SendMessageWithKeyboard(ctx, userCtx.UserID, text, rows)
		}
		return nil
	}
	return b.SendMessageWithKeyboard(ctx, userCtx.UserID, text, rows)
}

func (b *Bot) AnswerCallback(ctx context.Context, callbackID string, text string, rows [][]keyboards.Button) error {
	msgBody := &schemes.NewMessageBody{
		Text:   text,
		Format: "html",
	}

	if len(rows) > 0 {
		var buttonRows [][]schemes.ButtonInterface
		for _, row := range rows {
			var btnRow []schemes.ButtonInterface
			for _, btn := range row {
				var button schemes.ButtonInterface
				if btn.Type == "contact" {
					button = &schemes.RequestContactButton{
						Button: schemes.Button{Type: schemes.CONTACT, Text: btn.Text},
					}
				} else if btn.Type == "link" {
					button = &schemes.LinkButton{
						Button: schemes.Button{Type: schemes.LINK, Text: btn.Text},
						Url:    btn.Payload,
					}
				} else {
					button = &schemes.CallbackButton{
						Button:  schemes.Button{Type: schemes.CALLBACK, Text: btn.Text},
						Payload: btn.Payload,
						Intent:  schemes.POSITIVE,
					}
				}
				btnRow = append(btnRow, button)
			}
			if len(btnRow) > 0 {
				buttonRows = append(buttonRows, btnRow)
			}
		}
		if len(buttonRows) > 0 {
			keyboard := &schemes.Keyboard{Buttons: buttonRows}
			inlineKeyboard := &schemes.InlineKeyboardAttachmentRequest{
				AttachmentRequest: schemes.AttachmentRequest{Type: schemes.AttachmentKeyboard},
				Payload:           *keyboard,
			}
			msgBody.Attachments = append(msgBody.Attachments, inlineKeyboard)
		}
	}

	callbackAnswer := &schemes.CallbackAnswer{Message: msgBody}
	_, err := b.maxAPI.Messages.AnswerOnCallback(ctx, callbackID, callbackAnswer)
	if err != nil {
		log.Printf("AnswerCallback failed: %v", err)
		return err
	}
	return nil
}

func cleanPhone(s string) string {
	var result string
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') || c == '+' {
			result += string(c)
		}
	}
	return result
}

func extractPhoneFromVCF(vcfText string) string {
	if vcfText == "" {
		return ""
	}
	lines := splitLines(vcfText)
	for _, line := range lines {
		if len(line) > 3 && line[:3] == "TEL" {
			for i := 0; i < len(line); i++ {
				if line[i] == ':' && i < len(line)-1 {
					return cleanPhone(line[i+1:])
				}
			}
		}
	}
	return ""
}

func splitLines(s string) []string {
	var result []string
	var current string
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
			result = append(result, current)
			current = ""
			i++
		} else if s[i] == '\n' || s[i] == '\r' {
			result = append(result, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
