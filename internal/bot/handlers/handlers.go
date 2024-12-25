package handlers

import (
	"github.com/looplab/fsm"
	tele "gopkg.in/telebot.v4"
	"gorm.io/gorm"
)

const (
	StateStart       = "start"
	StateSelectBank  = "bank"
	StateEnterAmount = "amount"
	StateFinish      = "finish"

	// Админские состояния для добавления цели
	StateAddGoalTitle       = "add_goal_title"
	StateAddGoalDescription = "add_goal_description"
	StateAddGoalTargetSum   = "add_goal_target_sum"
	StateFinishedGoal       = "finished_goal"

	// Админские состояния для добавления администратора
	StateAddAdminID           = "add_admin_id"
	StateAddAdminWaitID       = "add_admin_wait_id"
	StateAddAdminWaitUsername = "add_admin_wait_username"

	// Админские состояния для редактирования цели
	StateEditGoalSelect      = "edit_goal_select"
	StateEditGoalFieldSelect = "edit_goal_field_select"
	StateEditGoalWaitInput   = "edit_goal_wait_input"
)

// Handler – базовая структура
type Handler struct {
	Bot               *tele.Bot
	DB                *gorm.DB
	AdminFSM          map[int64]*fsm.FSM
	UserFSM           map[int64]*fsm.FSM
	UserData          map[int64]map[string]interface{}
	SentMessages      map[int64][]tele.Message
	UserAlbumMessages map[int64][]*tele.Message
	GroupChatID       int64
}

// NewHandler – конструктор
func NewHandler(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}, groupChatID int64) *Handler {
	return &Handler{
		Bot:               bot,
		DB:                db,
		UserFSM:           userFSM,
		UserData:          userData,
		SentMessages:      make(map[int64][]tele.Message),
		UserAlbumMessages: make(map[int64][]*tele.Message),
		AdminFSM:          make(map[int64]*fsm.FSM),
		GroupChatID:       groupChatID,
	}
}

// RegisterHandlers – регистрация всех команд и кнопок
func RegisterHandlers(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}, groupChatID int64) {
	h := NewHandler(bot, db, userFSM, userData, groupChatID)

	// Регистрация команд
	bot.Handle("/start", h.onStart)
	bot.Handle("/panel", h.onPanel)
	bot.Handle("/add_admin", h.addFirstAdmin)
	bot.Handle("/goal", h.onGoalCommand)

	// Регистрация обработчиков сообщений
	bot.Handle(tele.OnContact, h.onContact)
	// bot.Handle(tele.OnText, h.onEnterAmount)
	bot.Handle(tele.OnPhoto, h.onUploadReceipt)
	bot.Handle(tele.OnText, h.onText)

	// Регистрация инлайн-кнопок
	registerInlineButtons(bot, h)
}

// registerInlineButtons – привязываем inline-кнопки к методам
func registerInlineButtons(bot *tele.Bot, h *Handler) {
	// Пример "пользовательских" кнопок
	bot.Handle(&tele.Btn{Unique: "donation"}, h.onDonation)
	bot.Handle(&tele.Btn{Unique: "info"}, h.onInfo)
	bot.Handle(&tele.Btn{Unique: "social"}, h.onSocial)
	bot.Handle(&tele.Btn{Unique: "goal"}, h.onGoal)

	// Админ-панель
	bot.Handle(&tele.Btn{Unique: "goals_panel"}, h.onGoalsPanel)
	bot.Handle(&tele.Btn{Unique: "add_goal"}, h.AddGoalHandler)
	bot.Handle(&tele.Btn{Unique: "priority_low"}, h.SetPriorityHandler)
	bot.Handle(&tele.Btn{Unique: "priority_medium"}, h.SetPriorityHandler)
	bot.Handle(&tele.Btn{Unique: "priority_high"}, h.SetPriorityHandler)

	bot.Handle(&tele.Btn{Unique: "statistic_panel"}, h.onStatistic)
	bot.Handle(&tele.Btn{Unique: "main_menu"}, h.onMainMenu)
	bot.Handle(&tele.Btn{Unique: "back_to_panel"}, h.onBackToPanel)
	bot.Handle(&tele.Btn{Unique: "add_admin"}, h.onAddAdminStart)

	bot.Handle(&tele.Btn{Unique: "list_goal"}, h.onListGoal)
	bot.Handle(&tele.Btn{Unique: "edit_goal"}, h.onEditGoal)
	bot.Handle(&tele.Btn{Unique: "edit_goal_select"}, h.onEditGoalSelect)
	bot.Handle(&tele.Btn{Unique: "edit_field"}, h.onEditField)
	bot.Handle(&tele.Btn{Unique: "edit_priority_select"}, h.onEditPrioritySelect)
	bot.Handle(&tele.Btn{Unique: "edit_status_select"}, h.onEditStatusSelect)
	bot.Handle(&tele.Btn{Unique: "delete_goal"}, h.onDeleteGoal)
	bot.Handle(&tele.Btn{Unique: "delete_goal_confirm"}, h.onDeleteGoalConfirm)
	bot.Handle(&tele.Btn{Unique: "edit_allocated_sum"}, h.onEditAllocatedSum)

	// Управление балансом
	bot.Handle(&tele.Btn{Unique: "balance_panel"}, h.onBalancePanel)
	bot.Handle(&tele.Btn{Unique: "add_funds"}, h.onAddFundsStart)
	bot.Handle(&tele.Btn{Unique: "sub_funds"}, h.onSubFundsStart)

	// Управление платежами (пример)
	bot.Handle(&tele.Btn{Unique: "sber"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "vtb"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "sbp"}, h.onBankDetails)

	// Кнопки «Назад»
	bot.Handle(&tele.Btn{Unique: "back"}, h.onBack)
	bot.Handle(&tele.Btn{Unique: "backAlbum"}, h.onBackAlbum)
}
