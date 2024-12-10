package handlers

import (
	tele "gopkg.in/telebot.v4"
	"gorm.io/gorm"
)

// Handler содержит зависимости для обработчиков
type Handler struct {
	Bot *tele.Bot
	DB  *gorm.DB
}

// RegisterHandlers регистрирует все обработчики
func RegisterHandlers(bot *tele.Bot, db *gorm.DB) {
	h := &Handler{Bot: bot, DB: db}

	// Команды
	bot.Handle("/start", h.onStart)
	bot.Handle(tele.OnContact, h.onContact)

	// Кнопки
	menu := &tele.ReplyMarkup{}
	btnDonation := menu.Data("🧡 Сделать пожертвование", "donation")
	btnInfo := menu.Data("ℹ️ Информация о фонде", "info")
	btnSocial := menu.Data("💬 Наши соц.сети", "social")
	btnGoal := menu.Data("🎯 Цели", "goal")

	bot.Handle(&btnDonation, h.onDonation)
	bot.Handle(&btnInfo, h.onInfo)
	bot.Handle(&btnSocial, h.onSocial)
	bot.Handle(&btnGoal, h.onGoal)

}
