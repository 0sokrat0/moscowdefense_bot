package handlers

import (
	tele "gopkg.in/telebot.v4"
	"log"
)

// onDonation обрабатывает кнопку "Сделать пожертвование"
func (h *Handler) onGoal(c tele.Context) error {
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	menu := &tele.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", "back") // Кнопка для возврата

	menu.Inline(
		menu.Row(btnBack), // Добавляем кнопку в меню
	)

	return c.Send("Спасибо за ваше желание помочь! Реквизиты для пожертвований: ...", menu)
}
