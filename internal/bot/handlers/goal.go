package handlers

import (
	"TgDonation/internal/database/models"
	"fmt"
	"log"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onGoal(c tele.Context) error {
	// Удаляем сообщение обратного вызова
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	// Загружаем активные цели из базы данных
	var goals []models.Goal
	if err := h.DB.Where("status = ?", "active").Find(&goals).Error; err != nil {
		return c.Send("Ошибка загрузки целей.")
	}

	// Если целей нет
	if len(goals) == 0 {
		return c.Send("Нет активных целей в данный момент.")
	}

	// Создаем меню с кнопками
	menu := &tele.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", "back")
	menu.Inline(
		menu.Row(btnBack),
	)

	// Формируем сообщение с информацией о целях
	message := "🎯 <b>Доступные цели:</b>\n"
	for i, g := range goals {
		progress := 0.0
		if g.TargetSum > 0 {
			progress = (g.CurrentSum / g.TargetSum) * 100
		}
		message += fmt.Sprintf("%d. <b>%s</b>\nСобрано: %.2f из %.2f (%.2f%%)\n\n", i+1, g.Title, g.CurrentSum, g.TargetSum, progress)
	}

	// Отправляем сообщение с клавиатурой
	return c.Send(message, &tele.SendOptions{
		ParseMode:   tele.ModeHTML, // Указываем режим парсинга HTML
		ReplyMarkup: menu,          // Передаем клавиатуру через SendOptions
	})
}
