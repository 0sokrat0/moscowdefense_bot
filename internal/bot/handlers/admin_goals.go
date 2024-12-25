package handlers

import (
	"context"
	"fmt"
	"github.com/looplab/fsm"
	"strconv"

	"TgDonation/internal/database/models"
	tele "gopkg.in/telebot.v4"
)

// Меню «Управление целями»
func (h *Handler) onGoalsPanel(c tele.Context) error {
	h.tryDeleteMessage(c)
	menu := &tele.ReplyMarkup{}
	AddGoal := menu.Data("➕ Добавить цель", "add_goal")
	ListGoal := menu.Data("📜 Список целей", "list_goal")
	EditGoal := menu.Data("✏️ Редактировать цель", "edit_goal")
	DeleteGoal := menu.Data("🗑️ Удалить цель", "delete_goal")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

	menu.Inline(
		menu.Row(AddGoal, ListGoal),
		menu.Row(EditGoal, DeleteGoal),
		menu.Row(BackBtn),
	)

	return c.Send("Управление целями:", menu)
}

// ------------------------ Добавить новую цель ------------------------

func (h *Handler) AddGoalHandler(c tele.Context) error {
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsmObj := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"action": "add_goal",
		"mode":   "admin",
	}

	// Переходим в состояние "add_goal_title"
	if err := fsmObj.Event(context.Background(), "add_goal_title"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка FSM. Начните процесс заново.")
	}

	return c.Send("Введите название новой цели:")
}

func (h *Handler) processGoalTitle(c tele.Context, fsmObj *fsm.FSM) error {
	title := c.Text()
	if title == "" {
		return c.Send("Название не может быть пустым.")
	}
	h.UserData[c.Sender().ID]["title"] = title

	if err := fsmObj.Event(context.Background(), "add_goal_description"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка FSM. Начните заново.")
	}
	return c.Send("Введите описание новой цели:")
}

func (h *Handler) processGoalDescription(c tele.Context, fsmObj *fsm.FSM) error {
	description := c.Text()
	if description == "" {
		return c.Send("Описание не может быть пустым.")
	}
	h.UserData[c.Sender().ID]["description"] = description

	if err := fsmObj.Event(context.Background(), "add_goal_target_sum"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка FSM. Начните заново.")
	}
	return c.Send("Введите целевую сумму (число):")
}

func (h *Handler) processGoalTargetSum(c tele.Context, fsmObj *fsm.FSM) error {
	targetSum, err := strconv.ParseFloat(c.Text(), 64)
	if err != nil || targetSum <= 0 {
		return c.Send("Некорректная сумма. Введите положительное число.")
	}
	h.UserData[c.Sender().ID]["target_sum"] = targetSum

	// Выбор приоритета
	menu := &tele.ReplyMarkup{}
	btnLow := menu.Data("🔵 Низкий", "priority_low", "low")
	btnMedium := menu.Data("🟠 Средний", "priority_medium", "medium")
	btnHigh := menu.Data("🔴 Высокий", "priority_high", "high")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
	menu.Inline(
		menu.Row(btnLow, btnMedium, btnHigh),
		menu.Row(BackBtn),
	)

	if err := fsmObj.Event(context.Background(), "finish_goal"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка FSM. Начните заново.")
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
		return c.Send("Некорректный приоритет.")
	}

	h.UserData[c.Sender().ID]["priority"] = priority

	// Создаём цель в БД
	title, _ := h.UserData[c.Sender().ID]["title"].(string)
	description, _ := h.UserData[c.Sender().ID]["description"].(string)
	targetSum, _ := h.UserData[c.Sender().ID]["target_sum"].(float64)

	goal := models.Goal{
		Title:       title,
		Description: description,
		TargetSum:   targetSum,
		Status:      "active",
		Priority:    priority,
		AdminID:     uint(c.Sender().ID),
	}

	if err := h.DB.Create(&goal).Error; err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка при сохранении цели. Попробуйте позже.")
	}

	// Завершаем
	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(back.Row(BackBtn))

	delete(h.UserData, c.Sender().ID)
	h.resetFSM(c.Sender().ID)

	return c.Send("✅ Цель успешно добавлена!\nВернитесь в админ-панель", back)
}

// ------------------------ Список целей ------------------------

func (h *Handler) onListGoal(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	var goals []models.Goal
	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
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
		message += fmt.Sprintf(
			"%d. <b>%s</b>\nID: %d\nОписание: %s\nЦелевая сумма: %.2f\nСтатус: %s\nПриоритет: %s\n\n",
			i+1, g.Title, g.ID, g.Description, g.TargetSum, g.Status, g.Priority,
		)
	}

	return c.Send(message, back, tele.ModeHTML)
}

// ------------------------ Редактирование цели ------------------------

func (h *Handler) onEditGoal(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsmObj := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"mode":   "admin",
		"action": "edit_goal",
	}

	if err := fsmObj.Event(context.Background(), "go_edit_goal_select"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Попробуйте снова.")
	}

	var goals []models.Goal
	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
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
		btn := menu.Data(
			fmt.Sprintf("%s (ID:%d)", g.Title, g.ID),
			"edit_goal_select",
			strconv.Itoa(int(g.ID)),
		)
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
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка при переходе к редактированию.")
	}

	h.UserData[c.Sender().ID]["goalID"] = goalID

	menu := &tele.ReplyMarkup{}
	btnTitle := menu.Data("Изменить название", "edit_field", "title")
	btnDesc := menu.Data("Изменить описание", "edit_field", "description")
	btnSum := menu.Data("Изменить целевую сумму", "edit_field", "target_sum")
	btnPriority := menu.Data("Изменить приоритет", "edit_field", "priority")

	// Вместо "Активна/Неактивна" — делаем только "Завершить цель"
	btnFinish := menu.Data("Завершить цель", "edit_status_select", "finished")

	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

	menu.Inline(
		menu.Row(btnTitle),
		menu.Row(btnDesc),
		menu.Row(btnSum),
		menu.Row(btnPriority),
		menu.Row(btnFinish),
		menu.Row(BackBtn),
	)

	return c.Edit(
		fmt.Sprintf("Редактирование цели: <b>%s</b>\nВыберите поле для редактирования:", goal.Title),
		menu,
		tele.ModeHTML,
	)
}

// Пример, если хотите редактировать AllocatedSum вручную (не всегда нужно)
func (h *Handler) onEditAllocatedSum(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	_, ok := h.UserData[c.Sender().ID]["goalID"].(int)
	if !ok {
		return c.Send("Ошибка: цель не выбрана.")
	}

	h.UserData[c.Sender().ID]["await_input"] = true
	return c.Send("Введите новую выделенную сумму для цели:")
}

func (h *Handler) processAllocatedSum(c tele.Context) error {
	if h.UserData[c.Sender().ID]["await_input"] != true {
		return c.Send("Нет активного запроса на ввод суммы.")
	}

	goalID, ok := h.UserData[c.Sender().ID]["goalID"].(int)
	if !ok {
		return c.Send("Ошибка: цель не выбрана.")
	}

	allocatedSumStr := c.Text()
	allocatedSum, err := strconv.ParseFloat(allocatedSumStr, 64)
	if err != nil || allocatedSum < 0 {
		return c.Send("Некорректная сумма. Введите положительное число.")
	}

	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		return c.Send("Цель не найдена.")
	}

	if err := h.DB.Save(&goal).Error; err != nil {
		return c.Send("Ошибка при сохранении данных цели.")
	}

	h.UserData[c.Sender().ID]["await_input"] = false
	return c.Send("✅ Выделенная сумма успешно обновлена!", backButton("back_to_panel"))
}

// Выбор поля для редактирования
func (h *Handler) onEditField(c tele.Context) error {
	field := c.Callback().Data
	h.UserData[c.Sender().ID]["edit_field"] = field

	adminFSM := h.getOrCreateAdminFSM(c.Sender().ID)

	if field == "title" || field == "description" || field == "target_sum" {
		// Переходим в состояние ожидания текстового ввода
		if err := adminFSM.Event(context.Background(), "wait_input"); err != nil {
			h.resetFSM(c.Sender().ID)
			return c.Send("Ошибка при ожидании ввода. Попробуйте снова.")
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
	}

	// Для статуса у нас теперь только кнопка "finished", обрабатывается на onEditGoalSelect
	return nil
}

// Обработка текстового ввода (title, description, target_sum)
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
			if newValue == "" {
				return c.Send("Описание не может быть пустым.")
			}
			goal.Description = newValue
		case "target_sum":
			val, err := strconv.ParseFloat(newValue, 64)
			if err != nil || val <= 0 {
				return c.Send("Некорректная целевая сумма. Введите положительное число.")
			}
			goal.TargetSum = val
		default:
			return c.Send("Неподдерживаемое поле для редактирования.")
		}

		if err := h.DB.Save(&goal).Error; err != nil {
			h.resetFSM(c.Sender().ID)
			return c.Send("Ошибка при сохранении цели.")
		}

		h.resetFSM(c.Sender().ID)
		return c.Send("✅ Цель успешно обновлена!\nВернитесь в админ-панель", backButton("back_to_panel"))
	}
	return nil
}

// Меняем приоритет (low / medium / high)
func (h *Handler) onEditPrioritySelect(c tele.Context) error {
	data := h.UserData[c.Sender().ID]
	if data == nil {
		return c.Send("Нет активного процесса редактирования.")
	}

	goalID, _ := data["goalID"].(int)
	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
	}

	priority := c.Callback().Data
	goal.Priority = priority
	if err := h.DB.Save(&goal).Error; err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при сохранении", ShowAlert: true})
	}
	h.resetFSM(c.Sender().ID)

	return c.Edit("✅ Приоритет успешно обновлён!\nВернитесь в админ-панель", backButton("back_to_panel"))
}

// Завершаем цель (finished) и вычитаем TargetSum
func (h *Handler) onEditStatusSelect(c tele.Context) error {
	data := h.UserData[c.Sender().ID]
	if data == nil {
		return c.Send("Нет активного процесса редактирования.")
	}
	goalID, _ := data["goalID"].(int)

	// Загружаем текущую цель
	var goal models.Goal
	if err := h.DB.First(&goal, goalID).Error; err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
	}

	oldStatus := goal.Status
	newStatus := c.Callback().Data // ожидаем "finished"

	// Ставим новый статус
	goal.Status = newStatus

	// Сохраняем
	if err := h.DB.Save(&goal).Error; err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при сохранении статуса", ShowAlert: true})
	}

	// Если впервые переводим в finished — вычитаем TargetSum из общего баланса
	if oldStatus != "finished" && newStatus == "finished" {
		// Списываем
		if err := h.subtractFromTotalDonation(goal.TargetSum); err != nil {
			// Если не удалось вычесть, откатываем
			goal.Status = oldStatus
			_ = h.DB.Save(&goal)

			h.resetFSM(c.Sender().ID)
			return c.Respond(&tele.CallbackResponse{
				Text:      "Недостаточно баланса или ошибка при списании.",
				ShowAlert: true,
			})
		}
	}

	h.resetFSM(c.Sender().ID)
	return c.Edit("✅ Цель завершена!\nВернитесь в админ-панель", backButton("back_to_panel"))
}

// Удаление цели
func (h *Handler) onDeleteGoal(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	var goals []models.Goal
	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
		return c.Send("Ошибка при загрузке целей.")
	}

	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")

	if len(goals) == 0 {
		back.Inline(back.Row(BackBtn))
		return c.Send("Нет целей для удаления.", back)
	}

	menu := &tele.ReplyMarkup{}
	rows := []tele.Row{}
	for _, g := range goals {
		btn := menu.Data(
			fmt.Sprintf("Удалить: %s (ID:%d)", g.Title, g.ID),
			"delete_goal_confirm",
			strconv.Itoa(int(g.ID)),
		)
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	back.Inline(menu.Row(BackBtn))

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

	// Удаляем (soft delete)
	if err := h.DB.Delete(&goal).Error; err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при удалении", ShowAlert: true})
	}

	// Или если хотите физически удалить: if err := h.DB.Unscoped().Delete(&goal).Error; ...

	return c.Edit("✅ Цель успешно удалена!\nВернитесь в админ-панель", backButton("back_to_panel"))
}
