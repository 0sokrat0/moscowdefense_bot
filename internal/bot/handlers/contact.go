package handlers

import (
	"TgDonation/internal/database/models"
	tele "gopkg.in/telebot.v4"
	"log"
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

	menu := &tele.ReplyMarkup{}
	btn1 := menu.Data("🧡 Сделать пожертвование", "donation")
	btn2 := menu.Data("ℹ️ Информация о фонде", "info")
	btn3 := menu.Data("📞 Связаться с нами", "contact")
	btn4 := menu.Data("🎯 Цели", "goal")

	menu.Inline(
		menu.Row(btn1),
		menu.Row(btn2),
		menu.Row(btn3),
		menu.Row(btn4),
	)

	photo := &tele.Photo{
		File:    tele.FromURL("https://disk.yandex.ru/i/ZTimPinmv7RioQ"),
		Caption: "<b>Добро пожаловать в бот \"Марфинский Тыл\"! 🇷🇺</b>\n\nМы помогаем укреплять тыл и поддерживать тех, кто участвует в СВО. Здесь вы можете:\n- Узнать, как помочь;\n- Сделать пожертвование;\n- Получить актуальную информацию о нашей работе.\n\n<b>Спасибо за вашу поддержку! Вместе мы сильнее.</b> 💪",
	}

	if _, err := c.Bot().Send(c.Chat(), photo, menu); err != nil {
		log.Printf("Ошибка отправки фото: %v", err)
		return err
	}

	return nil
}
