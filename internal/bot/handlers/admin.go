package handlers

import (
	"TgDonation/internal/database/models"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/looplab/fsm"
	tele "gopkg.in/telebot.v4"
)

// Админ-панель
func (h *Handler) onPanel(c tele.Context) error {
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к админ-панели.")
	}

	menu := &tele.ReplyMarkup{}
	GoalsBtn := menu.Data("Управление целями:", "goals_panel")
	StatisticBtn := menu.Data("Статистика", "statistic_panel")
	BroadcastBtn := menu.Data("Рассылка", "broadcast_panel")

	menu.Inline(
		menu.Row(GoalsBtn, StatisticBtn),
		menu.Row(BroadcastBtn),
	)

	return c.Send("Админ панель", menu)
}

// Кнопка "Назад в админ панель"
func (h *Handler) onBackToPanel(c tele.Context) error {
	// Пытаемся удалить текущее сообщение, если это callback
	if c.Callback() != nil && c.Callback().Message != nil {
		if err := c.Bot().Delete(c.Callback().Message); err != nil {
			log.Printf("Ошибка удаления сообщения: %v", err)
		}
	}
	return h.onPanel(c)
}

func (h *Handler) onGoalsPanel(c tele.Context) error {
	h.tryDeleteMessage(c)
	menu := &tele.ReplyMarkup{}
	AddGoal := menu.Data("Добавить цель", "add_goal")
	ListGoal := menu.Data("Список целей", "list_goal")
	EditGoal := menu.Data("Редактировать цель", "edit_goal")
	DeleteGoal := menu.Data("Удалить цель", "delete_goal")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

	menu.Inline(
		menu.Row(AddGoal, ListGoal),
		menu.Row(EditGoal, DeleteGoal),
		menu.Row(BackBtn),
	)

	return c.Send("Управление целями:", menu)
}

func (h *Handler) onStatisticPanel(c tele.Context) error {
	h.tryDeleteMessage(c)
	menu := &tele.ReplyMarkup{}
	Statistic := menu.Data("Статистика", "statistic")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

	menu.Inline(
		menu.Row(Statistic),
		menu.Row(BackBtn),
	)

	return c.Send("Статистика", menu)

}

func (h *Handler) onStatistic(c tele.Context) error {
	h.tryDeleteMessage(c)

	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	// Переменные для хранения результатов
	var totalDonations float64
	var donationsCount int64
	var activeGoalsCount int64
	var topGoal models.Goal

	// Общая сумма пожертвований
	if err := h.DB.Model(&models.Donation{}).Select("COALESCE(SUM(amount),0)").Scan(&totalDonations).Error; err != nil {
		log.Printf("Ошибка при подсчёте общей суммы пожертвований: %v", err)
		return c.Send("Ошибка при загрузке статистики.")
	}

	// Количество пожертвований
	if err := h.DB.Model(&models.Donation{}).Count(&donationsCount).Error; err != nil {
		log.Printf("Ошибка при подсчёте количества пожертвований: %v", err)
		return c.Send("Ошибка при загрузке статистики.")
	}

	// Средний размер пожертвования
	var avgDonation float64
	if donationsCount > 0 {
		avgDonation = totalDonations / float64(donationsCount)
	}

	// Количество активных целей
	if err := h.DB.Model(&models.Goal{}).Where("status = ?", "active").Count(&activeGoalsCount).Error; err != nil {
		log.Printf("Ошибка при подсчёте активных целей: %v", err)
		return c.Send("Ошибка при загрузке статистики.")
	}

	// Цель с наибольшей собранной суммой
	// Если целей нет, запрос вернёт ошибку или пустой результат
	if err := h.DB.Order("current_sum DESC").First(&topGoal).Error; err != nil {
		log.Printf("Ошибка при загрузке топ-цели: %v", err)
		// В случае отсутствия целей - не критичная ошибка, просто пропускаем
	}

	// Формируем текстовый отчет
	report := "<b>Статистика:</b>\n\n"
	report += fmt.Sprintf("💰 Общая сумма пожертвований: <b>%.2f</b>\n", totalDonations)
	report += fmt.Sprintf("📈 Количество пожертвований: <b>%d</b>\n", donationsCount)
	report += fmt.Sprintf("💲 Средний размер пожертвования: <b>%.2f</b>\n", avgDonation)
	report += fmt.Sprintf("🎯 Активных целей: <b>%d</b>\n", activeGoalsCount)

	if topGoal.ID != 0 {
		report += fmt.Sprintf("🏆 Топ цель по сбору: <b>%s</b> (%.2f из %.2f)\n", topGoal.Title, topGoal.CurrentSum, topGoal.TargetSum)
	} else {
		report += "🏆 Топ цель по сбору: Нет данных о целях\n"
	}

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))

	return c.Send(report, back, tele.ModeHTML)
}

// Начало добавления цели
func (h *Handler) AddGoalHandler(c tele.Context) error {
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"action": "add_goal",
		"mode":   "admin",
	}

	if err := fsm.Event(context.Background(), "add_goal_title"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка. Начните процесс заново.")
	}

	return c.Send("Введите название новой цели:")
}

func (h *Handler) processGoalTitle(c tele.Context, fsm *fsm.FSM) error {
	title := c.Text()
	if len(title) == 0 {
		return c.Send("Название не может быть пустым. Попробуйте снова.")
	}
	h.UserData[c.Sender().ID]["title"] = title
	if err := fsm.Event(context.Background(), "add_goal_description"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Начните процесс заново.")
	}
	return c.Send("Введите описание новой цели:")
}

func (h *Handler) processGoalDescription(c tele.Context, fsm *fsm.FSM) error {
	description := c.Text()
	if len(description) == 0 {
		return c.Send("Описание не может быть пустым. Попробуйте снова.")
	}
	h.UserData[c.Sender().ID]["description"] = description
	if err := fsm.Event(context.Background(), "add_goal_target_sum"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Начните процесс заново.")
	}
	return c.Send("Введите целевую сумму (число):")
}

func (h *Handler) processGoalTargetSum(c tele.Context, fsm *fsm.FSM) error {
	targetSum, err := strconv.ParseFloat(c.Text(), 64)
	if err != nil || targetSum <= 0 {
		return c.Send("Некорректная сумма. Введите положительное число.")
	}
	h.UserData[c.Sender().ID]["target_sum"] = targetSum

	menu := &tele.ReplyMarkup{}
	btnLow := menu.Data("🔵 Низкий", "priority_low", "low")
	btnMedium := menu.Data("🟠 Средний", "priority_medium", "medium")
	btnHigh := menu.Data("🔴 Высокий", "priority_high", "high")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
	menu.Inline(
		menu.Row(btnLow, btnMedium, btnHigh),
		menu.Row(BackBtn),
	)

	if err := fsm.Event(context.Background(), "finish_goal"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Начните заново.")
	}
	return c.Send("Выберите приоритет для этой цели:", menu)
}

func (h *Handler) SetPriorityHandler(c tele.Context) error {
	action, ok := h.UserData[c.Sender().ID]["action"]
	if !ok || action != "add_goal" {
		return c.Send("Неверный ввод. Попробуйте снова.")
	}

	priority := c.Callback().Data
	if priority != "low" && priority != "medium" && priority != "high" {
		return c.Send("Некорректный приоритет. Попробуйте снова.")
	}
	h.UserData[c.Sender().ID]["priority"] = priority

	title, _ := h.UserData[c.Sender().ID]["title"].(string)
	description, _ := h.UserData[c.Sender().ID]["description"].(string)
	targetSum, _ := h.UserData[c.Sender().ID]["target_sum"].(float64)

	goal := models.Goal{
		Title:       title,
		Description: description,
		TargetSum:   targetSum,
		CurrentSum:  0,
		Status:      "active",
		Priority:    priority,
		AdminID:     uint(c.Sender().ID),
	}

	if err := h.DB.Create(&goal).Error; err != nil {
		log.Printf("Ошибка при сохранении цели: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка при сохранении цели. Попробуйте позже.")
	}

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))

	delete(h.UserData, c.Sender().ID)
	h.resetFSM(c.Sender().ID)
	return c.Send("✅ Цель успешно добавлена!\nВернитесь в админ-панель", back)
}

func (h *Handler) onListGoal(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	var goals []models.Goal
	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
		log.Printf("Ошибка при загрузке целей: %v", err)
		return c.Send("Ошибка при загрузке целей.")
	}

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))

	if len(goals) == 0 {
		return c.Send("Целей нет.", back)
	}

	message := "🎯 <b>Список целей:</b>\n\n"
	for i, g := range goals {
		progress := 0.0
		if g.TargetSum > 0 {
			progress = (g.CurrentSum / g.TargetSum) * 100
		}
		message += fmt.Sprintf("%d. <b>%s</b> (ID: %d)\nОписание: %s\nСобрано: %.2f из %.2f (%.2f%%)\nСтатус: %s\nПриоритет: %s\n\n",
			i+1, g.Title, g.ID, g.Description, g.CurrentSum, g.TargetSum, progress, g.Status, g.Priority)
	}

	return c.Send(message, back, tele.ModeHTML)
}

// Редактирование целей
func (h *Handler) onEditGoal(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"mode":   "admin",
		"action": "edit_goal",
	}

	if err := fsm.Event(context.Background(), "go_edit_goal_select"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Попробуйте снова.")
	}

	var goals []models.Goal
	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
		log.Printf("Ошибка при загрузке целей: %v", err)
		return c.Send("Ошибка при загрузке целей.")
	}

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")

	if len(goals) == 0 {
		back.Inline(back.Row(BackBtn))
		return c.Send("Нет целей для редактирования.", back)
	}

	menu := &tele.ReplyMarkup{}
	rows := []tele.Row{}
	for _, g := range goals {
		btn := menu.Data(fmt.Sprintf("%s (ID:%d)", g.Title, g.ID), "edit_goal_select", strconv.Itoa(int(g.ID)))
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)

	return c.Send("Выберите цель для редактирования:", menu)
}

func (h *Handler) onEditGoalSelect(c tele.Context) error {
	goalIDStr := c.Callback().Data
	goalID, err := strconv.Atoi(goalIDStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Некорректный ID цели", ShowAlert: true})
	}

	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
	}

	adminFSM := h.getOrCreateAdminFSM(c.Sender().ID)
	if err := adminFSM.Event(context.Background(), "go_edit_goal_field"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка при переходе к редактированию. Попробуйте снова.")
	}

	h.UserData[c.Sender().ID]["goalID"] = goalID

	menu := &tele.ReplyMarkup{}
	btnTitle := menu.Data("Изменить название", "edit_field", "title")
	btnDesc := menu.Data("Изменить описание", "edit_field", "description")
	btnSum := menu.Data("Изменить целевую сумму", "edit_field", "target_sum")
	btnPriority := menu.Data("Изменить приоритет", "edit_field", "priority")
	btnStatus := menu.Data("Изменить статус", "edit_field", "status")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

	menu.Inline(
		menu.Row(btnTitle),
		menu.Row(btnDesc),
		menu.Row(btnSum),
		menu.Row(btnPriority),
		menu.Row(btnStatus),
		menu.Row(BackBtn),
	)

	return c.Edit(fmt.Sprintf("Редактирование цели: <b>%s</b>\nВыберите поле для редактирования:", goal.Title), menu, tele.ModeHTML)
}

func (h *Handler) onEditField(c tele.Context) error {

	field := c.Callback().Data
	h.UserData[c.Sender().ID]["edit_field"] = field

	adminFSM := h.getOrCreateAdminFSM(c.Sender().ID)

	if field == "title" || field == "description" || field == "target_sum" {
		if err := adminFSM.Event(context.Background(), "wait_input"); err != nil {
			log.Printf("FSM Event Error: %v", err)
			h.resetFSM(c.Sender().ID)
			return c.Send("Ошибка при установке состояния ожидания ввода. Попробуйте снова.")
		}
		h.UserData[c.Sender().ID]["await_input"] = true
		return c.Respond(&tele.CallbackResponse{Text: "Введите новое значение в чат"})
	}

	if field == "priority" {
		menu := &tele.ReplyMarkup{}
		btnLow := menu.Data("🔵 Низкий", "edit_priority_select", "low")
		btnMedium := menu.Data("🟠 Средний", "edit_priority_select", "medium")
		btnHigh := menu.Data("🔴 Высокий", "edit_priority_select", "high")
		BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

		menu.Inline(
			menu.Row(btnLow, btnMedium, btnHigh),
			menu.Row(BackBtn),
		)
		return c.Edit("Выберите новый приоритет:", menu)
	} else if field == "status" {
		menu := &tele.ReplyMarkup{}
		btnActive := menu.Data("Активна", "edit_status_select", "active")
		btnInactive := menu.Data("Неактивна", "edit_status_select", "inactive")
		BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

		menu.Inline(
			menu.Row(btnActive, btnInactive),
			menu.Row(BackBtn),
		)
		return c.Edit("Выберите новый статус:", menu)
	}

	return nil
}

func (h *Handler) onTextAdminEdit(c tele.Context, fsm *fsm.FSM) error {
	data := h.UserData[c.Sender().ID]
	if data == nil {
		return nil
	}
	if data["action"] == "edit_goal" && data["await_input"] == true && fsm.Current() == StateEditGoalWaitInput {
		goalID, _ := data["goalID"].(int)
		field, _ := data["edit_field"].(string)
		newValue := c.Text()

		var goal models.Goal
		if err := h.DB.First(&goal, goalID).Error; err != nil {
			log.Printf("Цель не найдена: %v", err)
			h.resetFSM(c.Sender().ID)
			return c.Send("Цель не найдена.")
		}

		switch field {
		case "title":
			if newValue == "" {
				return c.Send("Название не может быть пустым.")
			}
			goal.Title = newValue
		case "description":
			goal.Description = newValue
		case "target_sum":
			val, err := strconv.ParseFloat(newValue, 64)
			if err != nil || val <= 0 {
				return c.Send("Некорректная целевая сумма. Введите положительное число.")
			}
			goal.TargetSum = val
		}

		if err := h.DB.Save(&goal).Error; err != nil {
			log.Printf("Ошибка при сохранении цели: %v", err)
			h.resetFSM(c.Sender().ID)
			return c.Send("Ошибка при сохранении цели.")
		}

		h.resetFSM(c.Sender().ID)
		back := &tele.ReplyMarkup{}
		BackBtn := back.Data("⬅️ Назад", "back_to_panel")
		back.Inline(back.Row(BackBtn))
		return c.Send("✅ Цель успешно обновлена!\nВернитесь в админ-панель", back)
	}
	return nil
}

func (h *Handler) onEditPrioritySelect(c tele.Context) error {
	data := h.UserData[c.Sender().ID]
	if data == nil {
		return c.Send("Нет активного процесса редактирования.")
	}
	goalID, _ := data["goalID"].(int)
	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		log.Printf("Цель не найдена: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
	}

	priority := c.Callback().Data
	goal.Priority = priority
	if err := h.DB.Save(&goal).Error; err != nil {
		log.Printf("Ошибка при сохранении приоритета: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при сохранении", ShowAlert: true})
	}
	h.resetFSM(c.Sender().ID)

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))
	return c.Edit("✅ Цель успешно обновлена!\nВернитесь в админ-панель", back)
}

func (h *Handler) onEditStatusSelect(c tele.Context) error {
	data := h.UserData[c.Sender().ID]
	if data == nil {
		return c.Send("Нет активного процесса редактирования.")
	}
	goalID, _ := data["goalID"].(int)
	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		log.Printf("Цель не найдена: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
	}

	status := c.Callback().Data
	goal.Status = status
	if err := h.DB.Save(&goal).Error; err != nil {
		log.Printf("Ошибка при сохранении статуса: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при сохранении", ShowAlert: true})
	}
	h.resetFSM(c.Sender().ID)

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))
	return c.Edit("✅ Цель успешно обновлена!\nВернитесь в админ-панель", back)
}

// Удаление цели
func (h *Handler) onDeleteGoal(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	var goals []models.Goal
	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
		log.Printf("Ошибка при загрузке целей: %v", err)
		return c.Send("Ошибка при загрузке целей.")
	}

	menu := &tele.ReplyMarkup{}
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
	if len(goals) == 0 {
		menu.Inline(menu.Row(BackBtn))
		return c.Send("Нет целей для удаления.", menu)
	}

	rows := []tele.Row{}
	for _, g := range goals {
		btn := menu.Data(fmt.Sprintf("Удалить: %s (ID:%d)", g.Title, g.ID), "delete_goal_confirm", strconv.Itoa(int(g.ID)))
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)

	return c.Send("Выберите цель для удаления:", menu)
}

func (h *Handler) onDeleteGoalConfirm(c tele.Context) error {
	goalIDStr := c.Callback().Data
	goalID, err := strconv.Atoi(goalIDStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Некорректный ID цели", ShowAlert: true})
	}

	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
	}

	if err := h.DB.Delete(&goal).Error; err != nil {
		log.Printf("Ошибка при удалении цели: %v", err)
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при удалении", ShowAlert: true})
	}
	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))
	return c.Edit("✅ Цель успешно удалена!\nВернитесь в админ-панель", back)
}

// // Пересчёт сумм по целям
// func (h *Handler) onRecalcAllGoals(c tele.Context) error {
// 	var goals []models.Goal
// 	if err := h.DB.Find(&goals).Error; err != nil {
// 		log.Printf("Ошибка при загрузке целей: %v", err)
// 		return c.Send("Ошибка при загрузке целей.")
// 	}

// 	for _, g := range goals {
// 		if err := h.recalculateGoalCurrentSum(g.ID); err != nil {
// 			log.Printf("Ошибка при пересчёте цели %d: %v", g.ID, err)
// 		}
// 	}
// 	back := &tele.ReplyMarkup{}
// 	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
// 	back.Inline(back.Row(BackBtn))
// 	return c.Send("✅ Все цели пересчитаны.", back)
// }

// func (h *Handler) recalculateGoalCurrentSum(goalID uint) error {
// 	var total float64
// 	if err := h.DB.Model(&models.Donation{}).Where("goal_id = ?", goalID).Select("COALESCE(SUM(amount),0)").Scan(&total).Error; err != nil {
// 		return err
// 	}
// 	return h.DB.Model(&models.Goal{}).Where("id = ?", goalID).Update("current_sum", total).Error
// }
