package handlers

//const (
//	StateStart       = "start"
//	StateSelectBank  = "bank"
//	StateEnterAmount = "amount"
//	StateFinish      = "finish"
//	// Админские состояния для добавления цели
//	StateAddGoalTitle         = "add_goal_title"
//	StateAddGoalDescription   = "add_goal_description"
//	StateAddGoalTargetSum     = "add_goal_target_sum"
//	StateFinishedGoal         = "finished_goal"
//	StateAddAdminID           = "add_admin_id"
//	StateAddAdminWaitID       = "add_admin_wait_id"
//	StateAddAdminWaitUsername = "add_admin_wait_username"
//
//	// Админские состояния для редактирования цели
//	StateEditGoalSelect      = "edit_goal_select"
//	StateEditGoalFieldSelect = "edit_goal_field_select"
//	StateEditGoalWaitInput   = "edit_goal_wait_input"
//)
//
//func NewHandler(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}) *Handler {
//	return &Handler{
//		Bot:               bot,
//		DB:                db,
//		UserFSM:           userFSM,
//		UserData:          userData,
//		SentMessages:      make(map[int64][]tele.Message),
//		UserAlbumMessages: make(map[int64][]*tele.Message),
//		AdminFSM:          make(map[int64]*fsm.FSM),
//	}
//}
//
//func RegisterHandlers(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}) {
//	h := NewHandler(bot, db, userFSM, userData)
//
//	// Регистрация команд
//	bot.Handle("/start", h.onStart)
//	bot.Handle("/panel", h.onPanel)
//	bot.Handle("/add_admin", h.addFirstAdmin)
//	bot.Handle("/goal", h.onGoalCommand)
//
//	// Регистрация обработчиков сообщений
//	bot.Handle(tele.OnContact, h.onContact)
//	// bot.Handle(tele.OnText, h.onEnterAmount)
//	bot.Handle(tele.OnPhoto, h.onUploadReceipt)
//	bot.Handle(tele.OnText, h.onText)
//
//	// Регистрация кнопок
//	registerInlineButtons(bot, h)
//}
//
//func registerInlineButtons(bot *tele.Bot, h *Handler) {
//	//////////////////////////////////////////////////
//	// Пример "пользовательских" кнопок
//	bot.Handle(&tele.Btn{Unique: "donation"}, h.onDonation)
//	bot.Handle(&tele.Btn{Unique: "info"}, h.onInfo)
//	bot.Handle(&tele.Btn{Unique: "social"}, h.onSocial)
//	bot.Handle(&tele.Btn{Unique: "goal"}, h.onGoal)
//
//	//////////////////////////////////////////////////
//	// Админ-панель
//	bot.Handle(&tele.Btn{Unique: "goals_panel"}, h.onGoalsPanel)
//	bot.Handle(&tele.Btn{Unique: "add_goal"}, h.AddGoalHandler)
//	bot.Handle(&tele.Btn{Unique: "priority_low"}, h.SetPriorityHandler)
//	bot.Handle(&tele.Btn{Unique: "priority_medium"}, h.SetPriorityHandler)
//	bot.Handle(&tele.Btn{Unique: "priority_high"}, h.SetPriorityHandler)
//
//	bot.Handle(&tele.Btn{Unique: "statistic_panel"}, h.onStatisticPanel)
//	bot.Handle(&tele.Btn{Unique: "statistic"}, h.onStatistic)
//	bot.Handle(&tele.Btn{Unique: "main_menu"}, h.onMainMenu)
//	bot.Handle(&tele.Btn{Unique: "back_to_panel"}, h.onBackToPanel)
//	bot.Handle(&tele.Btn{Unique: "add_admin"}, h.onAddAdminStart)
//
//	bot.Handle(&tele.Btn{Unique: "list_goal"}, h.onListGoal)
//	bot.Handle(&tele.Btn{Unique: "edit_goal"}, h.onEditGoal)
//	bot.Handle(&tele.Btn{Unique: "edit_goal_select"}, h.onEditGoalSelect)
//	bot.Handle(&tele.Btn{Unique: "edit_field"}, h.onEditField)
//	bot.Handle(&tele.Btn{Unique: "edit_priority_select"}, h.onEditPrioritySelect)
//	bot.Handle(&tele.Btn{Unique: "edit_status_select"}, h.onEditStatusSelect)
//	bot.Handle(&tele.Btn{Unique: "delete_goal"}, h.onDeleteGoal)
//	bot.Handle(&tele.Btn{Unique: "delete_goal_confirm"}, h.onDeleteGoalConfirm)
//	bot.Handle(&tele.Btn{Unique: "edit_allocated_sum"}, h.onEditAllocatedSum)
//
//	// Управление балансом
//	bot.Handle(&tele.Btn{Unique: "balance_panel"}, h.onBalancePanel)
//	bot.Handle(&tele.Btn{Unique: "add_funds"}, h.onAddFundsStart)
//	bot.Handle(&tele.Btn{Unique: "sub_funds"}, h.onSubFundsStart)
//
//	//////////////////////////////////////////////////
//	// Управление платежами (пример)
//	bot.Handle(&tele.Btn{Unique: "sber"}, h.onBankDetails)
//	bot.Handle(&tele.Btn{Unique: "vtb"}, h.onBankDetails)
//	bot.Handle(&tele.Btn{Unique: "sbp"}, h.onBankDetails)
//
//	//////////////////////////////////////////////////
//	// Кнопки «Назад»
//	bot.Handle(&tele.Btn{Unique: "back"}, h.onBack)
//	bot.Handle(&tele.Btn{Unique: "backAlbum"}, h.onBackAlbum)
//}

////////////////////////////////////////////////////////////////////////////////
// FSM
////////////////////////////////////////////////////////////////////////////////

//// FSM для обычного пользователя (donation process)
//func (h *Handler) getOrCreateFSM(userID int64) *fsm.FSM {
//	if _, exists := h.UserFSM[userID]; !exists {
//		h.UserFSM[userID] = fsm.NewFSM(
//			StateStart,
//			fsm.Events{
//				{Name: "bank", Src: []string{StateStart}, Dst: StateSelectBank},
//				{Name: "amount", Src: []string{StateSelectBank}, Dst: StateEnterAmount},
//				{Name: "finish", Src: []string{StateEnterAmount}, Dst: StateFinish},
//			},
//			fsm.Callbacks{},
//		)
//	}
//	return h.UserFSM[userID]
//}
//
//// FSM для админа
//func (h *Handler) getOrCreateAdminFSM(userID int64) *fsm.FSM {
//	if _, exists := h.AdminFSM[userID]; !exists {
//		h.AdminFSM[userID] = fsm.NewFSM(
//			StateStart,
//			fsm.Events{
//				// Добавление цели
//				{Name: "add_goal_title", Src: []string{StateStart}, Dst: StateAddGoalTitle},
//				{Name: "add_goal_description", Src: []string{StateAddGoalTitle}, Dst: StateAddGoalDescription},
//				{Name: "add_goal_target_sum", Src: []string{StateAddGoalDescription}, Dst: StateAddGoalTargetSum},
//				{Name: "finish_goal", Src: []string{StateAddGoalTargetSum}, Dst: StateFinishedGoal},
//
//				// Добавление администратора
//				{Name: "add_admin_id", Src: []string{StateStart}, Dst: StateAddAdminWaitID},
//				{Name: "add_admin_username", Src: []string{StateAddAdminWaitID}, Dst: StateAddAdminWaitUsername},
//				{Name: "finish_add_admin", Src: []string{StateAddAdminWaitUsername}, Dst: StateFinish},
//
//				// Редактирование цели
//				{Name: "go_edit_goal_select", Src: []string{StateStart}, Dst: StateEditGoalSelect},
//				{Name: "go_edit_goal_field", Src: []string{StateEditGoalSelect}, Dst: StateEditGoalFieldSelect},
//				{Name: "wait_input", Src: []string{StateEditGoalFieldSelect}, Dst: StateEditGoalWaitInput},
//			},
//			fsm.Callbacks{},
//		)
//	}
//	return h.AdminFSM[userID]
//}

////////////////////////////////////////////////////////////////////////////////
// Обработчик текстовых сообщений: смотрим режим (admin/user) и текущее состояние
////////////////////////////////////////////////////////////////////////////////

//func (h *Handler) onText(c tele.Context) error {
//	mode, _ := h.UserData[c.Sender().ID]["mode"].(string)
//
//	// Если админ
//	if mode == "admin" {
//		if adminFSM, exists := h.AdminFSM[c.Sender().ID]; exists {
//			switch adminFSM.Current() {
//			// Добавление новой цели
//			case StateAddGoalTitle:
//				return h.processGoalTitle(c, adminFSM)
//			case StateAddGoalDescription:
//				return h.processGoalDescription(c, adminFSM)
//			case StateAddGoalTargetSum:
//				return h.processGoalTargetSum(c, adminFSM)
//
//			// Если мы в «добавлении средств»
//			case "add_funds":
//				return h.processAddFunds(c, adminFSM)
//			case "sub_funds":
//				return h.processSubFunds(c, adminFSM)
//
//			// Редактирование цели: ввод текстовых значений (title, desc, target_sum)
//			case StateEditGoalWaitInput:
//				return h.onTextAdminEdit(c, adminFSM)
//
//			default:
//				log.Printf("[WARN] Неизвестное состояние FSM администратора ID=%d: %s", c.Sender().ID, adminFSM.Current())
//				h.resetFSM(c.Sender().ID)
//				return c.Send("Неверный ввод. Начните заново или используйте /panel.")
//			}
//		} else {
//			h.resetFSM(c.Sender().ID)
//			return c.Send("Нет активного админского процесса. Попробуйте снова.")
//		}
//	}
//
//	// Если обычный пользователь
//	if mode == "user" {
//		if userFSM, exists := h.UserFSM[c.Sender().ID]; exists {
//			switch userFSM.Current() {
//			case StateSelectBank:
//				return h.onBankDetails(c)
//			case StateEnterAmount:
//				return h.onEnterAmount(c)
//			default:
//				log.Printf("[WARN] Неизвестное состояние FSM пользователя ID=%d: %s", c.Sender().ID, userFSM.Current())
//				h.resetFSM(c.Sender().ID)
//				return c.Send("Неверный ввод. Начните заново.")
//			}
//		} else {
//			h.resetFSM(c.Sender().ID)
//			return c.Send("Нет активного процесса. Введите /start, чтобы начать.")
//		}
//	}
//
//	// Если режим не установлен
//	h.resetFSM(c.Sender().ID)
//	return c.Send("У вас нет активного процесса. Введите /start, чтобы начать.")
//}
