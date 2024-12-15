package handlers

import (
	"TgDonation/internal/database/models"
	"log"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onBack(c tele.Context) error {
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	return h.onStart(c)
}

func (h *Handler) isAdminFromDB(tgID int) bool {
	var count int64
	h.DB.Model(&models.Admin{}).Where("tg_id = ?", tgID).Count(&count)
	return count > 0
}

func (h *Handler) addFirstAdmin(c tele.Context) error {
	var count int64
	h.DB.Model(&models.Admin{}).Count(&count)

	if count > 0 {
		return c.Send("Администратор уже существует. Вы не можете использовать эту команду.")
	}

	admin := models.Admin{
		TgID:     c.Sender().ID,
		Username: c.Sender().Username,
		Role:     "superadmin",
	}

	if err := h.DB.Create(&admin).Error; err != nil {
		log.Printf("Ошибка при добавлении администратора: %v", err)
		return c.Send("Ошибка при добавлении администратора.")
	}

	return c.Send("Вы успешно добавлены как администратор.")
}

func (h *Handler) tryDeleteMessage(c tele.Context) {
	if c.Callback() != nil && c.Callback().Message != nil {
		if err := c.Bot().Delete(c.Callback().Message); err != nil {
			log.Printf("Ошибка удаления сообщения: %v", err)
		}
	}
}

// func (h *Handler) tryDeleteOnText(c tele.Context) {
// 	if c.Message().Text != "" {
// 		if err := c.Bot().Delete(c.Message()); err != nil {
// 			log.Printf("Ошибка удаления сообщения: %v", err)
// 		}
// 	}
// }

func (h *Handler) deleteUserMessage(c tele.Context) error {
	msg := c.Message()
	if msg == nil {
		// Если сообщение пустое (например, был callback), удалить нечего
		return nil
	}

	// Пытаемся удалить сообщение пользователя
	err := c.Bot().Delete(msg)
	if err != nil {
		log.Printf("Ошибка при удалении сообщения пользователя: %v", err)
	}
	return err
}

// func (h *Handler) editPreviousMessage(c tele.Context, newText string, markup *tele.ReplyMarkup) error {
// 	var msg *tele.Message

// 	// Если вызывается из callback-хэндлера, сообщение обычно находится в c.Callback().Message
// 	if c.Callback() != nil && c.Callback().Message != nil {
// 		msg = c.Callback().Message
// 	} else if c.Message() != nil {
// 		// Если вызывается из хэндлера на обычное сообщение, то используем c.Message()
// 		msg = c.Message()
// 	}

// 	if msg == nil {
// 		return fmt.Errorf("нет сообщения для редактирования")
// 	}

// 	// Попытка отредактировать сообщение
// 	_, err := c.Bot().Edit(msg, newText, markup, tele.ModeHTML)
// 	if err != nil {
// 		log.Printf("Ошибка при редактировании сообщения: %v", err)
// 	}

// 	return err
// }
