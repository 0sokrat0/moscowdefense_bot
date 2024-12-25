package handlers

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strings"

	"TgDonation/internal/database/models"
	tele "gopkg.in/telebot.v4"
)

func getPriorityInRussian(priority string) string {
	switch strings.ToLower(priority) {
	case "high":
		return "Высокий"
	case "medium":
		return "Средний"
	case "low":
		return "Низкий"
	default:
		return priority // Если приоритет неизвестен, возвращаем как есть
	}
}

// Вспомогательная функция для убирания лишних нулей после запятой
func formatFloatNoTrailingZeros(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// Обработчик для отображения целей и общего баланса
func (h *Handler) onGoal(c tele.Context) error {
	// Удаляем старое сообщение (если это Callback)
	if c.Callback() != nil {
		if err := c.Bot().Delete(c.Callback().Message); err != nil {
			log.Printf("Ошибка удаления сообщения: %v", err)
		}
	}

	// 1) Получаем общий баланс из TotalDonation
	var totalRec models.TotalDonation
	err := h.DB.First(&totalRec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Если записи нет, создаём её с нулевым балансом
			totalRec = models.TotalDonation{
				Total: 0,
			}
			if err := h.DB.Create(&totalRec).Error; err != nil {
				log.Printf("Ошибка создания записи TotalDonation: %v", err)
				return c.Send("Ошибка при загрузке общего баланса.")
			}
		} else {
			log.Printf("Ошибка при получении TotalDonation: %v", err)
			return c.Send("Ошибка при загрузке общего баланса.")
		}
	}

	// 2) Загружаем активные цели
	var goals []models.Goal
	if err := h.DB.Where("status = ?", "active").Find(&goals).Error; err != nil {
		return c.Send("Ошибка загрузки целей.")
	}

	// 3) Создаём меню (кнопка «Назад»)
	menu := &tele.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", "back_to_panel")
	menu.Inline(menu.Row(btnBack))

	// 4) Формируем сообщение
	// Начинаем с общего баланса
	message := fmt.Sprintf(
		"<b>💰 Общий баланс боевой копилки:</b> %s ₽\n",
		formatFloatNoTrailingZeros(totalRec.Total),
	)
	message += "--------------------------------\n"

	// Проверяем наличие активных целей
	if len(goals) == 0 {
		message += "<b>🎯 Активные цели:</b> Нет активных целей в данный момент.\n"
	} else {
		message += "<b>🎯 Активные цели:</b>\n\n"
		for i, g := range goals {
			message += fmt.Sprintf(
				"%d. <b>%s</b> (ID: %d)\n   📄 Описание: %s\n   🎯 Целевая сумма: %s ₽\n   🔺 Приоритет: %s\n\n",
				i+1,
				g.Title,
				g.ID,
				g.Description,
				formatFloatNoTrailingZeros(g.TargetSum),
				getPriorityInRussian(g.Priority), // Преобразуем первый символ в верхний регистр
			)
		}
	}

	// 5) Отправляем сообщение с меню
	return c.Send(message, &tele.SendOptions{
		ParseMode:   tele.ModeHTML,
		ReplyMarkup: menu,
	})
}

func (h *Handler) onGoalCommand(c tele.Context) error {
	// 1) Получаем общий баланс из TotalDonation
	var totalRec models.TotalDonation
	err := h.DB.First(&totalRec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Если записи нет, создаём её с нулевым балансом
			totalRec = models.TotalDonation{
				Total: 0,
			}
			if err := h.DB.Create(&totalRec).Error; err != nil {
				log.Printf("Ошибка создания записи TotalDonation: %v", err)
				return c.Send("Ошибка при загрузке общего баланса.")
			}
		} else {
			log.Printf("Ошибка при получении TotalDonation: %v", err)
			return c.Send("Ошибка при загрузке общего баланса.")
		}
	}

	// 2) Загружаем активные цели
	var goals []models.Goal
	if err := h.DB.Where("status = ?", "active").Find(&goals).Error; err != nil {
		return c.Send("Ошибка загрузки целей.")
	}

	// 3) Создаём меню (кнопка «Назад»)
	menu := &tele.ReplyMarkup{}
	btnBack := menu.URL("⬅️ Вернуться в бот", "https://t.me/moscowdefense_bot?start")
	menu.Inline(menu.Row(btnBack))

	// 4) Формируем сообщение
	message := fmt.Sprintf(
		"<b>💰 Общий баланс боевой копилки:</b> %s ₽\n",
		formatFloatNoTrailingZeros(totalRec.Total),
	)
	message += "--------------------------------\n"

	if len(goals) == 0 {
		message += "<b>🎯 Активные цели:</b> Нет активных целей в данный момент.\n"
	} else {
		message += "<b>🎯 Активные цели:</b>\n\n"
		for i, g := range goals {
			message += fmt.Sprintf(
				"%d. <b>%s</b>\n   📄 Описание: %s\n   🎯 Целевая сумма: %s ₽\n   🔺 Приоритет: %s\n\n",
				i+1,
				g.Title,
				g.Description,
				formatFloatNoTrailingZeros(g.TargetSum),
				getPriorityInRussian(g.Priority),
			)
		}
	}

	// 5) Отправляем сообщение с меню
	return c.Send(message, &tele.SendOptions{
		ParseMode:   tele.ModeHTML,
		ReplyMarkup: menu,
	})
}
