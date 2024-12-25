package handlers

import (
	"TgDonation/internal/database/models"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onContact(c tele.Context) error {
	contact := c.Message().Contact
	if contact == nil {
		return c.Send("Контактная информация не предоставлена.")
	}

	user := models.User{
		TgID:        contact.UserID,
		Username:    c.Sender().Username,
		PhoneNumber: contact.PhoneNumber,
	}

	// Сохранение нового пользователя
	if err := h.DB.Create(&user).Error; err != nil {
		return c.Send("Ошибка при сохранении данных пользователя.")
	}

	removeKeyboard := &tele.ReplyMarkup{
		RemoveKeyboard: true,
	}

	reaction := tele.Reaction{
		Type:  "emoji",
		Emoji: "👀",
	}

	reactions := tele.Reactions{
		Reactions: []tele.Reaction{reaction},
		Big:       false,
	}

	if err := c.Bot().React(c.Sender(), c.Message(), reactions); err != nil {
		return c.Send("Не удалось добавить реакцию.")
	}

	// Отправляем сообщение с удалением клавиатуры
	if err := c.Send("✅", removeKeyboard); err != nil {

	}

	return h.onStart(c)
}
