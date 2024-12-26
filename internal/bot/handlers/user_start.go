package handlers

import (
	"TgDonation/internal/database/models"
	"log"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onStart(c tele.Context) error {

	h.deleteUserMessage(c)

	if c.Chat().Type != tele.ChatPrivate {
		botmenu := &tele.ReplyMarkup{}
		shareBtn := botmenu.URL("Перейти в бота", "t.me/moscowdefense_bot")
		botmenu.Inline(
			botmenu.Row(shareBtn),
		)
		return c.Send("Пожалуйста, перейдите в личный чат с ботом для использования этой команды", botmenu)
	}

	var existingUser models.User
	if err := h.DB.Where("tg_id = ?", c.Sender().ID).First(&existingUser).Error; err != nil {
		menu := &tele.ReplyMarkup{
			IsPersistent:    false,
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		}
		btnRequestContact := menu.Contact("📱 Отправить контакт")

		menu.Reply(
			menu.Row(btnRequestContact),
		)

		return c.Send("Пожалуйста, отправьте ваш контакт для регистрации.", menu)
	}
	menu := &tele.ReplyMarkup{}
	btn1 := menu.Data("🧡 Сделать пожертвование", "donation")
	btn2 := menu.Data("ℹ️ Информация о фонде", "info")
	btn3 := menu.Data("💬 Наши соц.сети", "social")
	btn4 := menu.Data("🎯 Цели", "goal")

	menu.Inline(
		menu.Row(btn1),
		menu.Row(btn4),
		menu.Row(btn2),
		menu.Row(btn3),
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
