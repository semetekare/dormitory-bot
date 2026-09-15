package templates

import "fmt"

const (
	MsgWelcome = "🎓 Привет! Я бот твоего общежития.\n\nЯ помогу тебе:\n👕 Записаться на стирку\n🚪 Посмотреть мат. ответственность\n📖 Прочитать справки и правила\n💬 Найти чаты общежития\n🧹 Узнать график дежурств\n\nДля начала работы мне нужен твой номер телефона."

	MsgAskContact = "📲 Пожалуйста, отправь свой номер телефона для авторизации."

	MsgVerificationInProgress = "⏳ Проверяю твои данные в базе университета..."

	MsgVerificationSuccess = "✅ Ты успешно верифицирован!\n\n%s"

	MsgVerificationFailed = "❌ Не удалось найти тебя в базе университета.\n\nПроверь правильность номера телефона или обратись в администрацию общежития."

	MsgEISUnavailable = "⚠️ Сервис проверки временно недоступен. Попробуй позже."

	MsgChooseRole = "Кто ты?\n\nВыбери свою роль:"

	MsgResidentData = "🏠 Твоё общежитие: %s\n🚪 Твоя комната: %s, этаж %d\n\nВсё верно?"

	MsgResidentNotFound = "😔 Твоё общежитие не найдено в системе.\n\nОбратись в администрацию общежития для регистрации."

	MsgResidentConfirmed = "✅ Отлично! Ты подтверждён как житель.\n\nДобро пожаловать в главное меню!"

	MsgEmployeeInit = "🔍 Ищу информацию о тебе как о сотруднике..."

	MsgEmployeeNotFound = "😔 Не удалось найти информацию о тебе как о сотруднике.\n\nОбратись к администратору системы."

	MsgSelectDormitory = "🏢 Выбери общежитие для работы:"

	MsgAuthorizedEmployee = "👔 %s, ваша должность — %s."

	MsgMainMenu = "📋 Главное меню. Выбери раздел:"

	MsgLaundryViewQueue = "👕 Доступные слоты для стирки на %s:\n\n%s"

	MsgLaundryNoSlots = "🕐 На сегодня свободных слотов нет.\nПопробуй завтра или позже."

	MsgLaundryBookSuccess = "✅ Ты записан на стирку!\n\nМашина: %s\nДата: %s\nВремя: %s - %s"

	MsgLaundryCancelConfirm = "⚠️ Ты уверен, что хочешь отменить запись на стирку?\n\nМашина: %s\nДата: %s\nВремя: %s - %s"

	MsgLaundryCancelSuccess = "✅ Запись на стирку отменена."

	MsgRoomItems = "🚪 Твоя комната: %s\n\n📋 Материальная ответственность:\n\n%s"

	MsgRoomNoItems = "📋 Список мат. ответственности пуст."

	MsgInfoCategories = "📖 Выбери категорию:"

	MsgInfoItem = "📖 %s\n\n%s"

	MsgChatLinks = "💬 Ссылки на чаты:\n\n%s"

	MsgChatLinksEmpty = "💬 Ссылок на чаты пока нет."

	MsgCleaningSchedule = "🧹 График дежурств твоей комнаты на %s:\n\n%s"

	MsgCleaningEmpty = "🧹 Дежурств на этот месяц нет."

	MsgBack = "Возвращаемся назад..."

	MsgUnknownCommand = "🤔 Я пока не понимаю эту команду. Выбери действие в меню ниже."

	MsgGlobalError = "⚠️ Упс! Возникла техническая проблема. Попробуй повторить запрос позже."

	MsgGenericError = "❌ Произошла ошибка. Попробуйте позже или обратитесь к администратору."

	MsgFSMError = "🔄 Произошёл сбой. Отправь /start чтобы начать заново."
)

func FormatResidentData(dormName, roomNumber string, floor int) string {
	return fmt.Sprintf(MsgResidentData, dormName, roomNumber, floor)
}

func FormatVerificationSuccess(fullName string) string {
	return fmt.Sprintf(MsgVerificationSuccess, fullName)
}

func FormatAuthorizedEmployee(fullName, role string) string {
	return fmt.Sprintf(MsgAuthorizedEmployee, fullName, role)
}

func FormatLaundrySlot(machineNumber string, startTime, endTime string) string {
	return fmt.Sprintf("🅿️ Машина %s: %s - %s", machineNumber, startTime, endTime)
}

func FormatRoomItem(itemName, invNumber string, quantity int, condition string) string {
	return fmt.Sprintf("• %s\n  Инв. №: %s\n  Кол-во: %d\n  Состояние: %s", itemName, invNumber, quantity, condition)
}

func FormatChatLink(title, url, platform string) string {
	return fmt.Sprintf("• %s (%s)\n  %s", title, platform, url)
}

func FormatCleaningDuty(date, dutyType string) string {
	typeEmoji := "🧹"
	switch dutyType {
	case "penalty":
		typeEmoji = "⚠️"
	case "kitchen":
		typeEmoji = "🍳"
	case "garbage":
		typeEmoji = "🗑️"
	}
	return fmt.Sprintf("%s %s - %s", typeEmoji, date, dutyType)
}
