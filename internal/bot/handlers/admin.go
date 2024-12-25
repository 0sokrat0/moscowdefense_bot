package handlers

//// Создание кнопки "Назад"
//func backButton(data string) *tele.ReplyMarkup {
//	back := &tele.ReplyMarkup{}
//	BackBtn := back.Data("⬅️ Назад", data)
//	back.Inline(back.Row(BackBtn))
//	return back
//}

// Админ-панель
//func (h *Handler) onPanel(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к админ-панели.")
//	}
//
//	menu := &tele.ReplyMarkup{}
//	GoalsBtn := menu.Data("📝 Управление целями", "goals_panel")
//	StatisticBtn := menu.Data("📊 Статистика", "statistic_panel")
//	BroadcastBtn := menu.Data("📨 Рассылка", "broadcast_panel")
//	AllocateFundsBtn := menu.Data("🔄 Распределить средства", "allocate_funds")
//	AddAdminBtn := menu.Data("➕ Добавить администратора", "add_admin")
//	BalanceBtn := menu.Data("💰 Управление балансом", "balance_panel")
//	BackBtn := menu.Data("🔙 Назад", "back")
//
//	menu.Inline(
//		menu.Row(GoalsBtn, StatisticBtn),
//		menu.Row(BroadcastBtn),
//		menu.Row(AllocateFundsBtn),
//		menu.Row(AddAdminBtn),
//		menu.Row(BalanceBtn),
//		menu.Row(BackBtn),
//	)
//
//	return c.Send("Админ панель", menu)
//}

//func (h *Handler) onBalancePanel(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	menu := &tele.ReplyMarkup{}
//	AddFunds := menu.Data("➕ Добавить средства", "add_funds")
//	SubFunds := menu.Data("➖ Вычесть средства", "sub_funds")
//	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
//
//	menu.Inline(
//		menu.Row(AddFunds),
//		menu.Row(SubFunds),
//		menu.Row(BackBtn),
//	)
//
//	return c.Send("Управление общим балансом:", menu)
//}
//
//func (h *Handler) onAddFundsStart(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	h.resetFSM(c.Sender().ID)
//	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
//	h.UserData[c.Sender().ID] = map[string]interface{}{
//		"action": "add_funds",
//		"mode":   "admin",
//	}
//
//	if err := fsm.Event(context.Background(), "wait_add_funds_amount"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	return c.Send("Введите сумму для добавления к общему балансу:")
//}
//
//func (h *Handler) processAddFunds(c tele.Context, fsm *fsm.FSM) error {
//	amountStr := c.Text()
//	amount, err := strconv.ParseFloat(amountStr, 64)
//	if err != nil || amount <= 0 {
//		return c.Send("Некорректная сумма. Введите положительное число.")
//	}
//
//	if err := h.addToTotalDonation(amount); err != nil {
//		log.Printf("Ошибка при добавлении средств: %v", err)
//		return c.Send("Ошибка при добавлении средств.")
//	}
//
//	// Перераспределяем средства после изменения общего баланса
//	if err := h.reallocateFundsForAllGoals(); err != nil {
//		log.Printf("Ошибка при перераспределении средств: %v", err)
//	}
//
//	h.resetFSM(c.Sender().ID)
//	delete(h.UserData, c.Sender().ID)
//	return c.Send(fmt.Sprintf("✅ Добавлено %.2f к общему балансу", amount), backButton("back_to_panel"))
//}
//
//func (h *Handler) onSubFundsStart(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	h.resetFSM(c.Sender().ID)
//	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
//	h.UserData[c.Sender().ID] = map[string]interface{}{
//		"action": "sub_funds",
//		"mode":   "admin",
//	}
//
//	if err := fsm.Event(context.Background(), "wait_sub_funds_amount"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	return c.Send("Введите сумму для вычета из общего баланса:")
//}
//
//func (h *Handler) processSubFunds(c tele.Context, fsm *fsm.FSM) error {
//	amountStr := c.Text()
//	amount, err := strconv.ParseFloat(amountStr, 64)
//	if err != nil || amount <= 0 {
//		return c.Send("Некорректная сумма. Введите положительное число.")
//	}
//
//	if err := h.subtractFromTotalDonation(amount); err != nil {
//		log.Printf("Ошибка при вычитании средств: %v", err)
//		return c.Send("Ошибка при вычитании средств.")
//	}
//
//	// Перераспределяем средства после изменения общего баланса
//	if err := h.reallocateFundsForAllGoals(); err != nil {
//		log.Printf("Ошибка при перераспределении средств: %v", err)
//	}
//
//	h.resetFSM(c.Sender().ID)
//	delete(h.UserData, c.Sender().ID)
//	return c.Send(fmt.Sprintf("✅ Вычтено %.2f из общего баланса", amount), backButton("back_to_panel"))
//}
//
//func (h *Handler) addToTotalDonation(amount float64) error {
//	var totalRec models.TotalDonation
//	err := h.DB.First(&totalRec).Error
//	if err != nil {
//		// Если записи нет, создаём
//		totalRec.Total = amount
//		return h.DB.Create(&totalRec).Error
//	}
//	totalRec.Total += amount
//	return h.DB.Save(&totalRec).Error
//}
//
//func (h *Handler) subtractFromTotalDonation(amount float64) error {
//	var totalRec models.TotalDonation
//	err := h.DB.First(&totalRec).Error
//	if err != nil {
//		// Нет записей или ошибка – тогда нечего вычитать
//		return fmt.Errorf("нет доступных средств для вычета")
//	}
//	if totalRec.Total < amount {
//		// Если пытаемся вычесть больше, чем есть, уменьшим до нуля
//		amount = totalRec.Total
//	}
//	totalRec.Total -= amount
//	return h.DB.Save(&totalRec).Error
//}

//func (h *Handler) onAddAdminStart(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	h.resetFSM(c.Sender().ID)
//	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
//	h.UserData[c.Sender().ID] = map[string]interface{}{
//		"action": "add_admin",
//		"mode":   "admin",
//	}
//
//	if err := fsm.Event(context.Background(), "add_admin_id"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	return c.Send("Введите TG ID нового администратора (числом):")
//}
//
//func (h *Handler) processNewAdminID(c tele.Context, fsm *fsm.FSM) error {
//	idStr := c.Text()
//	newAdminID, err := strconv.Atoi(idStr)
//	if err != nil || newAdminID <= 0 {
//		return c.Send("Некорректный ID. Введите положительное число.")
//	}
//
//	h.UserData[c.Sender().ID]["new_admin_id"] = newAdminID
//
//	if err := fsm.Event(context.Background(), "add_admin_username"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	return c.Send("Введите username нового администратора (без @):")
//}
//
//func (h *Handler) processNewAdminUsername(c tele.Context, fsm *fsm.FSM) error {
//	username := c.Text()
//	if username == "" {
//		return c.Send("Username не может быть пустым.")
//	}
//
//	h.UserData[c.Sender().ID]["new_admin_username"] = username
//
//	if err := fsm.Event(context.Background(), "finish_add_admin"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка при добавлении администратора. Попробуйте снова.")
//	}
//
//	return h.finishAddAdmin(c)
//}
//
//func (h *Handler) finishAddAdmin(c tele.Context) error {
//	data := h.UserData[c.Sender().ID]
//	if data == nil {
//		return c.Send("Нет данных для добавления администратора.")
//	}
//
//	newAdminID := data["new_admin_id"].(int)
//	newAdminUsername := data["new_admin_username"].(string)
//
//	var count int64
//	h.DB.Model(&models.Admin{}).Where("tg_id = ?", newAdminID).Count(&count)
//	if count > 0 {
//		h.resetFSM(c.Sender().ID)
//		delete(h.UserData, c.Sender().ID)
//		return c.Send("Администратор с таким TG ID уже существует.", backButton("back_to_panel"))
//	}
//
//	admin := models.Admin{
//		TgID:     int64(newAdminID),
//		Username: newAdminUsername,
//		Role:     "admin",
//	}
//
//	if err := h.DB.Create(&admin).Error; err != nil {
//		h.resetFSM(c.Sender().ID)
//		delete(h.UserData, c.Sender().ID)
//		return c.Send("Ошибка при добавлении администратора.", backButton("back_to_panel"))
//	}
//
//	h.resetFSM(c.Sender().ID)
//	delete(h.UserData, c.Sender().ID)
//	return c.Send("✅ Администратор успешно добавлен!", backButton("back_to_panel"))
//}

//func (h *Handler) onBackToPanel(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	return h.onPanel(c)
//}

//func (h *Handler) onGoalsPanel(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	menu := &tele.ReplyMarkup{}
//	AddGoal := menu.Data("➕ Добавить цель", "add_goal")
//	ListGoal := menu.Data("📜 Список целей", "list_goal")
//	EditGoal := menu.Data("✏️ Редактировать цель", "edit_goal")
//	DeleteGoal := menu.Data("🗑️ Удалить цель", "delete_goal")
//	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
//
//	menu.Inline(
//		menu.Row(AddGoal, ListGoal),
//		menu.Row(EditGoal, DeleteGoal),
//		menu.Row(BackBtn),
//	)
//
//	return c.Send("Управление целями:", menu)
//}

//func (h *Handler) onStatisticPanel(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	menu := &tele.ReplyMarkup{}
//	Statistic := menu.Data("🧮 Статистика", "statistic")
//	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
//
//	menu.Inline(
//		menu.Row(Statistic),
//		menu.Row(BackBtn),
//	)
//
//	return c.Send("Статистика", menu)
//}
//
//func (h *Handler) onStatistic(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	var totalDonations float64
//	var donationsCount int64
//	var activeGoalsCount int64
//	var topGoal models.Goal
//
//	if err := h.DB.Model(&models.Donation{}).Select("COALESCE(SUM(amount),0)").Scan(&totalDonations).Error; err != nil {
//		return c.Send("Ошибка при загрузке статистики.")
//	}
//
//	if err := h.DB.Model(&models.Donation{}).Count(&donationsCount).Error; err != nil {
//		return c.Send("Ошибка при загрузке статистики.")
//	}
//
//	var avgDonation float64
//	if donationsCount > 0 {
//		avgDonation = totalDonations / float64(donationsCount)
//	}
//
//	if err := h.DB.Model(&models.Goal{}).Where("status = ?", "active").Count(&activeGoalsCount).Error; err != nil {
//		return c.Send("Ошибка при загрузке статистики.")
//	}
//
//	if err := h.DB.Order("current_sum DESC").First(&topGoal).Error; err != nil {
//		// Если целей нет - просто пропускаем
//	}
//
//	report := "<b>Статистика:</b>\n\n"
//	report += fmt.Sprintf("💰 Общая сумма пожертвований: <b>%.2f</b>\n", totalDonations)
//	report += fmt.Sprintf("📈 Количество пожертвований: <b>%d</b>\n", donationsCount)
//	report += fmt.Sprintf("💲 Средний размер пожертвования: <b>%.2f</b>\n", avgDonation)
//	report += fmt.Sprintf("🎯 Активных целей: <b>%d</b>\n", activeGoalsCount)
//
//	if topGoal.ID != 0 {
//		report += fmt.Sprintf("🏆 Топ цель по сбору: <b>%s</b> (Целевая: %.2f)\n", topGoal.Title, topGoal.TargetSum)
//	} else {
//		report += "🏆 Топ цель по сбору: Нет данных\n"
//	}
//
//	back := backButton("back_to_panel")
//	return c.Send(report, back, tele.ModeHTML)
//}

//func (h *Handler) onAllocateFunds(c tele.Context) error {
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	h.resetFSM(c.Sender().ID)
//	f := h.getOrCreateAdminFSM(c.Sender().ID)
//	if err := f.Event(context.Background(), "start_allocate"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	h.UserData[c.Sender().ID] = map[string]interface{}{
//		"action": "allocate_funds",
//		"mode":   "admin",
//	}
//
//	var goals []models.Goal
//	if err := h.DB.Where("deleted_at IS NULL AND status <> 'finished'").Find(&goals).Error; err != nil {
//		return c.Send("Ошибка при загрузке целей.")
//	}
//
//	if len(goals) == 0 {
//		return c.Send("Нет целей для распределения средств.")
//	}
//
//	response := "Доступные цели:\n"
//	for _, g := range goals {
//		response += fmt.Sprintf("ID: %d | Название: %s | Целевая сумма: %.2f | Выделено: %.2f\n", g.ID, g.Title, g.TargetSum, g.AllocatedSum)
//	}
//	response += "\nВведите ID цели, на которую хотите выделить средства:"
//
//	return c.Send(response)
//}
//
//func (h *Handler) processAllocateGoalSelect(c tele.Context, fsm *fsm.FSM) error {
//	goalIDStr := c.Text()
//	goalID, err := strconv.Atoi(goalIDStr)
//	if err != nil {
//		return c.Send("Некорректный ID. Попробуйте снова.")
//	}
//
//	var goal models.Goal
//	if err := h.DB.First(&goal, goalID).Error; err != nil {
//		return c.Send("Цель не найдена.")
//	}
//
//	h.UserData[c.Sender().ID]["allocate_goal_id"] = goalID
//
//	if err := fsm.Event(context.Background(), "allocate_wait_sum"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	return c.Send(fmt.Sprintf("Цель: %s\nВведите сумму для выделения:", goal.Title))
//}
//
//func (h *Handler) processAllocateSum(c tele.Context, fsm *fsm.FSM) error {
//	data := h.UserData[c.Sender().ID]
//	if data == nil {
//		return c.Send("Нет данных для распределения.")
//	}
//	goalID, _ := data["allocate_goal_id"].(int)
//
//	sumStr := c.Text()
//	allocateSum, err := strconv.ParseFloat(sumStr, 64)
//	if err != nil || allocateSum <= 0 {
//		return c.Send("Некорректная сумма. Введите положительное число.")
//	}
//
//	free, err := h.getFreeFunds()
//	if err != nil {
//		return c.Send("Ошибка при вычислении свободных средств.")
//	}
//
//	if allocateSum > free {
//		allocateSum = free
//	}
//
//	if allocateSum == 0 {
//		h.resetFSM(c.Sender().ID)
//		delete(h.UserData, c.Sender().ID)
//		return c.Send("Недостаточно средств для выделения.", backButton("back_to_panel"))
//	}
//
//	var goal models.Goal
//	if err := h.DB.First(&goal, goalID).Error; err != nil {
//		h.resetFSM(c.Sender().ID)
//		delete(h.UserData, c.Sender().ID)
//		return c.Send("Цель не найдена.", backButton("back_to_panel"))
//	}
//
//	goal.AllocatedSum += allocateSum
//	if err := h.DB.Save(&goal).Error; err != nil {
//		h.resetFSM(c.Sender().ID)
//		delete(h.UserData, c.Sender().ID)
//		return c.Send("Ошибка при выделении средств цели.", backButton("back_to_panel"))
//	}
//
//	// После выделения средств попробуем снова перераспределить, чтобы поддержать целостность
//	if err := h.reallocateFundsForAllGoals(); err != nil {
//		log.Printf("Ошибка перераспределения средств: %v", err)
//	}
//
//	h.resetFSM(c.Sender().ID)
//	delete(h.UserData, c.Sender().ID)
//	return c.Send(fmt.Sprintf("✅ Выделено %.2f для цели \"%s\"", allocateSum, goal.Title), backButton("back_to_panel"))
//}
//
//func (h *Handler) getFreeFunds() (float64, error) {
//	var total float64
//	if err := h.DB.Model(&models.TotalDonation{}).Select("COALESCE(SUM(total),0)").Scan(&total).Error; err != nil {
//		return 0, err
//	}
//
//	var goals []models.Goal
//	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
//		return 0, err
//	}
//
//	var alreadyAllocated float64
//	for _, g := range goals {
//		alreadyAllocated += g.AllocatedSum
//	}
//
//	free := total - alreadyAllocated
//	if free < 0 {
//		free = 0
//	}
//	return free, nil
//}
//
//func (h *Handler) reallocateFundsForAllGoals() error {
//	var total float64
//	if err := h.DB.Model(&models.TotalDonation{}).Select("COALESCE(SUM(total),0)").Scan(&total).Error; err != nil {
//		return err
//	}
//
//	var goals []models.Goal
//	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
//		return err
//	}
//
//	var alreadyAllocated float64
//	for _, g := range goals {
//		alreadyAllocated += g.AllocatedSum
//	}
//
//	free := total - alreadyAllocated
//	if free <= 0 {
//		return nil
//	}
//
//	for i := range goals {
//		goal := &goals[i]
//		if goal.TargetSum > goal.AllocatedSum && goal.Status != "finished" {
//			needed := goal.TargetSum - goal.AllocatedSum
//			toAllocate := needed
//			if toAllocate > free {
//				toAllocate = free
//			}
//			goal.AllocatedSum += toAllocate
//			free -= toAllocate
//
//			if err := h.DB.Save(goal).Error; err != nil {
//				return fmt.Errorf("ошибка при обновлении цели %d: %v", goal.ID, err)
//			}
//
//			if free <= 0 {
//				break
//			}
//		}
//	}
//
//	return nil
//}

//func (h *Handler) AddGoalHandler(c tele.Context) error {
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	h.resetFSM(c.Sender().ID)
//	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
//	h.UserData[c.Sender().ID] = map[string]interface{}{
//		"action": "add_goal",
//		"mode":   "admin",
//	}
//
//	if err := fsm.Event(context.Background(), "add_goal_title"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Произошла ошибка. Начните процесс заново.")
//	}
//
//	return c.Send("Введите название новой цели:")
//}
//
//func (h *Handler) processGoalTitle(c tele.Context, fsm *fsm.FSM) error {
//	title := c.Text()
//	if title == "" {
//		return c.Send("Название не может быть пустым.")
//	}
//	h.UserData[c.Sender().ID]["title"] = title
//	if err := fsm.Event(context.Background(), "add_goal_description"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Начните процесс заново.")
//	}
//	return c.Send("Введите описание новой цели:")
//}
//
//func (h *Handler) processGoalDescription(c tele.Context, fsm *fsm.FSM) error {
//	description := c.Text()
//	if description == "" {
//		return c.Send("Описание не может быть пустым.")
//	}
//	h.UserData[c.Sender().ID]["description"] = description
//	if err := fsm.Event(context.Background(), "add_goal_target_sum"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Начните процесс заново.")
//	}
//	return c.Send("Введите целевую сумму (число):")
//}
//
//func (h *Handler) processGoalTargetSum(c tele.Context, fsm *fsm.FSM) error {
//	targetSum, err := strconv.ParseFloat(c.Text(), 64)
//	if err != nil || targetSum <= 0 {
//		return c.Send("Некорректная сумма. Введите положительное число.")
//	}
//	h.UserData[c.Sender().ID]["target_sum"] = targetSum
//
//	menu := &tele.ReplyMarkup{}
//	btnLow := menu.Data("🔵 Низкий", "priority_low", "low")
//	btnMedium := menu.Data("🟠 Средний", "priority_medium", "medium")
//	btnHigh := menu.Data("🔴 Высокий", "priority_high", "high")
//	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
//	menu.Inline(
//		menu.Row(btnLow, btnMedium, btnHigh),
//		menu.Row(BackBtn),
//	)
//
//	if err := fsm.Event(context.Background(), "finish_goal"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Начните заново.")
//	}
//	return c.Send("Выберите приоритет для этой цели:", menu)
//}
//
//func (h *Handler) SetPriorityHandler(c tele.Context) error {
//	action, ok := h.UserData[c.Sender().ID]["action"]
//	if !ok || action != "add_goal" {
//		return c.Send("Неверный ввод. Попробуйте снова.")
//	}
//
//	priority := c.Callback().Data
//	if priority != "low" && priority != "medium" && priority != "high" {
//		return c.Send("Некорректный приоритет.")
//	}
//	h.UserData[c.Sender().ID]["priority"] = priority
//
//	title, _ := h.UserData[c.Sender().ID]["title"].(string)
//	description, _ := h.UserData[c.Sender().ID]["description"].(string)
//	targetSum, _ := h.UserData[c.Sender().ID]["target_sum"].(float64)
//
//	goal := models.Goal{
//		Title:        title,
//		Description:  description,
//		TargetSum:    targetSum,
//		Status:       "active",
//		Priority:     priority,
//		AdminID:      uint(c.Sender().ID),
//		AllocatedSum: 0,
//	}
//
//	if err := h.DB.Create(&goal).Error; err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка при сохранении цели. Попробуйте позже.")
//	}
//
//	if err := h.reallocateFundsForAllGoals(); err != nil {
//		log.Printf("Ошибка при выделении средств для новой цели: %v", err)
//	}
//
//	back := &tele.ReplyMarkup{}
//	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
//	back.Inline(back.Row(BackBtn))
//
//	delete(h.UserData, c.Sender().ID)
//	h.resetFSM(c.Sender().ID)
//	return c.Send("✅ Цель успешно добавлена!\nВернитесь в админ-панель", back)
//}
//
//func (h *Handler) onListGoal(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	var goals []models.Goal
//	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
//		return c.Send("Ошибка при загрузке целей.")
//	}
//
//	back := &tele.ReplyMarkup{}
//	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
//	back.Inline(back.Row(BackBtn))
//
//	if len(goals) == 0 {
//		return c.Send("Целей нет.", back)
//	}
//
//	message := "🎯 <b>Список целей:</b>\n\n"
//	for i, g := range goals {
//		message += fmt.Sprintf("%d. <b>%s</b>\nID: %d\nОписание: %s\nЦелевая сумма: %.2f\nВыделено: %.2f\nСтатус: %s\nПриоритет: %s\n\n",
//			i+1, g.Title, g.ID, g.Description, g.TargetSum, g.AllocatedSum, g.Status, g.Priority)
//	}
//
//	return c.Send(message, back, tele.ModeHTML)
//}
//
//func (h *Handler) onEditGoal(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	h.resetFSM(c.Sender().ID)
//	fsm := h.getOrCreateAdminFSM(c.Sender().ID)
//	h.UserData[c.Sender().ID] = map[string]interface{}{
//		"mode":   "admin",
//		"action": "edit_goal",
//	}
//
//	if err := fsm.Event(context.Background(), "go_edit_goal_select"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка. Попробуйте снова.")
//	}
//
//	var goals []models.Goal
//	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
//		return c.Send("Ошибка при загрузке целей.")
//	}
//
//	back := &tele.ReplyMarkup{}
//	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
//
//	if len(goals) == 0 {
//		back.Inline(back.Row(BackBtn))
//		return c.Send("Нет целей для редактирования.", back)
//	}
//
//	menu := &tele.ReplyMarkup{}
//	rows := []tele.Row{}
//	for _, g := range goals {
//		btn := menu.Data(fmt.Sprintf("%s (ID:%d)", g.Title, g.ID), "edit_goal_select", strconv.Itoa(int(g.ID)))
//		rows = append(rows, menu.Row(btn))
//	}
//	menu.Inline(rows...)
//
//	return c.Send("Выберите цель для редактирования:", menu)
//}
//
//func (h *Handler) onEditGoalSelect(c tele.Context) error {
//	goalIDStr := c.Callback().Data
//	goalID, err := strconv.Atoi(goalIDStr)
//	if err != nil {
//		return c.Respond(&tele.CallbackResponse{Text: "Некорректный ID цели", ShowAlert: true})
//	}
//
//	var goal models.Goal
//	if err := h.DB.First(&goal, goalID).Error; err != nil {
//		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
//	}
//
//	adminFSM := h.getOrCreateAdminFSM(c.Sender().ID)
//	if err := adminFSM.Event(context.Background(), "go_edit_goal_field"); err != nil {
//		h.resetFSM(c.Sender().ID)
//		return c.Send("Ошибка при переходе к редактированию.")
//	}
//
//	h.UserData[c.Sender().ID]["goalID"] = goalID
//
//	menu := &tele.ReplyMarkup{}
//	btnTitle := menu.Data("Изменить название", "edit_field", "title")
//	btnDesc := menu.Data("Изменить описание", "edit_field", "description")
//	btnSum := menu.Data("Изменить целевую сумму", "edit_field", "target_sum")
//	btnGoalSum := menu.Data("Изменить выделенную сумму", "edit_allocated_sum", "allocated_sum")
//	btnPriority := menu.Data("Изменить приоритет", "edit_field", "priority")
//	btnStatus := menu.Data("Изменить статус", "edit_field", "status")
//	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")
//
//	menu.Inline(
//		menu.Row(btnTitle),
//		menu.Row(btnDesc),
//		menu.Row(btnSum),
//		menu.Row(btnGoalSum),
//		menu.Row(btnPriority),
//		menu.Row(btnStatus),
//		menu.Row(BackBtn),
//	)
//
//	return c.Edit(fmt.Sprintf("Редактирование цели: <b>%s</b>\nВыберите поле для редактирования:", goal.Title), menu, tele.ModeHTML)
//}

// Удаление цели
//func (h *Handler) onDeleteGoal(c tele.Context) error {
//	h.tryDeleteMessage(c)
//	if !h.isAdminFromDB(int(c.Sender().ID)) {
//		return c.Send("У вас нет доступа к этой функции.")
//	}
//
//	var goals []models.Goal
//	if err := h.DB.Where("deleted_at IS NULL").Find(&goals).Error; err != nil {
//		return c.Send("Ошибка при загрузке целей.")
//	}
//
//	back := &tele.ReplyMarkup{}
//	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
//
//	if len(goals) == 0 {
//		back.Inline(back.Row(BackBtn))
//		return c.Send("Нет целей для удаления.", back)
//	}
//
//	menu := &tele.ReplyMarkup{}
//	rows := []tele.Row{}
//	for _, g := range goals {
//		btn := menu.Data(fmt.Sprintf("Удалить: %s (ID:%d)", g.Title, g.ID), "delete_goal_confirm", strconv.Itoa(int(g.ID)))
//		rows = append(rows, menu.Row(btn))
//	}
//	menu.Inline(rows...)
//	back.Inline(menu.Row(BackBtn))
//
//	return c.Send("Выберите цель для удаления:", menu)
//}
//
//func (h *Handler) onDeleteGoalConfirm(c tele.Context) error {
//	goalIDStr := c.Callback().Data
//	goalID, err := strconv.Atoi(goalIDStr)
//	if err != nil {
//		return c.Respond(&tele.CallbackResponse{Text: "Некорректный ID цели", ShowAlert: true})
//	}
//
//	var goal models.Goal
//	if err := h.DB.First(&goal, goalID).Error; err != nil {
//		return c.Respond(&tele.CallbackResponse{Text: "Цель не найдена", ShowAlert: true})
//	}
//
//	if err := h.DB.Delete(&goal).Error; err != nil {
//		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при удалении", ShowAlert: true})
//	}
//
//	if err := h.reallocateFundsForAllGoals(); err != nil {
//		log.Printf("Ошибка при перераспределении средств после удаления цели: %v", err)
//	}
//
//	return c.Edit("✅ Цель успешно удалена!\nВернитесь в админ-панель", backButton("back_to_panel"))
//}
