package keyboards

import (
	"fmt"
	"strings"
)

const (
	PayloadAuthRequest      = "auth:request"
	PayloadMenuLaundry      = "menu:laundry"
	PayloadMenuRoom         = "menu:room"
	PayloadMenuInfo         = "menu:info"
	PayloadMenuReferences   = "menu:references"
	PayloadModuleRefs       = "module_refs"
	PayloadMenuChatLinks    = "menu:chat_links"
	PayloadMenuCleaning     = "menu:cleaning"
	PayloadNavBack          = "nav:back"
	PayloadNavHome          = "nav:home"
	PayloadResidentConfirm  = "resident:confirm"
	PayloadResidentReject   = "resident:reject"
	PayloadEmployeeInit     = "employee:init"
	PayloadLaundryViewQueue = "laundry:view_queue"
	PayloadLaundryBook      = "laundry:book"
	PayloadLaundryCancel    = "laundry:cancel"
	PayloadRoomViewItems    = "room:view_items"
	PayloadInfoViewList     = "info:view_list"
	PayloadCleaningView     = "cleaning:view"
)

func PayloadLaundryBookSlot(machineID, startTime string) string {
	return fmt.Sprintf("laundry:book:%s:%s", machineID, startTime)
}

func PayloadLaundryCancelBooking(bookingID string) string {
	return fmt.Sprintf("laundry:cancel:%s", bookingID)
}

func PayloadInfoViewItem(materialID string) string {
	return fmt.Sprintf("info:view:%s", materialID)
}

func PayloadSelectDormitory(dormitoryID string) string {
	return fmt.Sprintf("dormitory:select:%s", dormitoryID)
}

type Button struct {
	Text    string
	Payload string
	Type    string
}

func GetMainMenuKeyboard() [][]Button {
	return GetMainMenuKeyboardWithRoles(nil)
}

func GetMainMenuKeyboardWithRoles(roles []string) [][]Button {
	rows := [][]Button{
		{{Text: "👕 Стирка", Payload: PayloadMenuLaundry}},
		{{Text: "🚪 Моя комната", Payload: PayloadMenuRoom}},
		{{Text: "📖 Справки", Payload: PayloadMenuReferences}},
		{{Text: "💬 Чаты общежития", Payload: PayloadMenuChatLinks}},
		{{Text: "🧹 Дежурство", Payload: PayloadMenuCleaning}},
	}

	if len(roles) > 0 {
		roleLabel := "🔑 Роль"
		for _, r := range roles {
			roleLabel = fmt.Sprintf("🔑 Роль: %s", r)
			break
		}
		rows = append(rows, []Button{{Text: roleLabel, Payload: "resident_role:select"}})
	}

	return rows
}

func GetContactKeyboard() [][]Button {
	return [][]Button{
		{{Text: "📱 Отправить контакт", Payload: "", Type: "contact"}},
	}
}

func GetConfirmationKeyboard(confirmPayload, cancelPayload string) [][]Button {
	return [][]Button{
		{{Text: "✅ Да", Payload: confirmPayload}, {Text: "❌ Нет", Payload: cancelPayload}},
	}
}

func GetBackKeyboard() [][]Button {
	return [][]Button{
		{{Text: "⬅️ Назад", Payload: PayloadNavBack}},
	}
}

func GetHomeKeyboard() [][]Button {
	return [][]Button{
		{{Text: "🏠 На главную", Payload: PayloadNavHome}},
	}
}

func GetDormitorySelectKeyboard(dormitories []struct{ ID, Name string }) [][]Button {
	rows := make([][]Button, 0, len(dormitories)+1)
	for _, d := range dormitories {
		rows = append(rows, []Button{{
			Text:    d.Name,
			Payload: PayloadSelectDormitory(d.ID),
		}})
	}
	rows = append(rows, []Button{{Text: "⬅️ Назад", Payload: PayloadNavBack}})
	return rows
}

func GetAdminMenuKeyboard(role string) [][]Button {
	var rows [][]Button
	if role == "director" || role == "commandant" {
		rows = append(rows, []Button{{Text: "🤖 Модули ботов", Payload: "menu:modules"}})
	}
	rows = append(rows, []Button{{Text: "👕 Управление прачкой", Payload: "menu:laundry_mgmt"}})
	if role == "director" || role == "commandant" || role == "chairman" {
		rows = append(rows, []Button{{Text: "🚪 Комнаты и жильцы", Payload: "menu:room_mgmt"}})
	}
	if role == "director" || role == "commandant" || role == "chairman" || role == "starosta" {
		rows = append(rows, []Button{{Text: "🧹 Дежурства", Payload: "menu:cleaning_mgmt"}})
	}
	rows = append(rows, []Button{{Text: "💬 Ссылки на чаты", Payload: "menu:chat_links_mgmt"}})
	if role == "director" || role == "commandant" {
		rows = append(rows, []Button{{Text: "📖 Справки", Payload: "menu:reference_mgmt"}})
	}
	return rows
}

func GetRoleSelectKeyboard() [][]Button {
	return [][]Button{
		{{Text: "🏠 Житель", Payload: "role:resident"}},
		{{Text: "👔 Сотрудник", Payload: "role:employee"}},
	}
}

func GetModuleSelectKeyboard() [][]Button {
	return [][]Button{
		{{Text: "👕 Прачка", Payload: "module:laundry"}, {Text: "🚪 Комната", Payload: "module:my_room"}},
		{{Text: "🧹 Дежурства", Payload: "module:cleaning"}},
		{{Text: "💬 Чаты", Payload: "module:chat_links"}, {Text: "📖 Общие", Payload: "module:references"}},
		{{Text: "⬅️ Назад", Payload: PayloadNavBack}},
	}
}

func GetPlatformSelectKeyboard() [][]Button {
	return [][]Button{
		{{Text: "💬 MAX", Payload: "platform:max"}},
		{{Text: "💬 ВКонтакте", Payload: "platform:vk"}},
		{{Text: "💬 Свой вариант", Payload: "platform:other"}},
		{{Text: "⬅️ Назад", Payload: PayloadNavBack}},
	}
}

func FormatMAXMention(platformUserID, fullName string) string {
	if strings.HasPrefix(platformUserID, "eis_") {
		return ""
	}
	return fmt.Sprintf("💬 Написать в MAX: %s (ID: %s)", fullName, platformUserID)
}
