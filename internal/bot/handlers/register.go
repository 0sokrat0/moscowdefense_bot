package handlers

import (
	"log"

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

	// Админские состояния для редактирования цели
	StateEditGoalSelect      = "edit_goal_select"
	StateEditGoalFieldSelect = "edit_goal_field_select"
	StateEditGoalWaitInput   = "edit_goal_wait_input"
)

type Handler struct {
	Bot               *tele.Bot
	DB                *gorm.DB
	AdminFSM          map[int64]*fsm.FSM
	UserFSM           map[int64]*fsm.FSM
	UserData          map[int64]map[string]interface{}
	SentMessages      map[int64][]tele.Message
	UserAlbumMessages map[int64][]*tele.Message
}

func NewHandler(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}) *Handler {
	return &Handler{
		Bot:               bot,
		DB:                db,
		UserFSM:           userFSM,
		UserData:          userData,
		SentMessages:      make(map[int64][]tele.Message),
		UserAlbumMessages: make(map[int64][]*tele.Message),
		AdminFSM:          make(map[int64]*fsm.FSM),
	}
}

func RegisterHandlers(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}) {
	h := NewHandler(bot, db, userFSM, userData)

	// Регистрация команд
	bot.Handle("/start", h.onStart)
	bot.Handle("/panel", h.onPanel)
	bot.Handle("/add_admin", h.addFirstAdmin)

	// Регистрация обработчиков сообщений
	bot.Handle(tele.OnContact, h.onContact)
	// bot.Handle(tele.OnText, h.onEnterAmount)
	bot.Handle(tele.OnPhoto, h.onUploadReceipt)
	bot.Handle(tele.OnText, h.onText)

	// Регистрация кнопок
	registerInlineButtons(bot, h)
}

func registerInlineButtons(bot *tele.Bot, h *Handler) {

	// Привязка кнопок
	bot.Handle(&tele.Btn{Unique: "donation"}, h.onDonation)
	bot.Handle(&tele.Btn{Unique: "info"}, h.onInfo)
	bot.Handle(&tele.Btn{Unique: "social"}, h.onSocial)
	bot.Handle(&tele.Btn{Unique: "goal"}, h.onGoal)

	// Админ панель
	bot.Handle(&tele.Btn{Unique: "goals_panel"}, h.onGoalsPanel)

	bot.Handle(&tele.Btn{Unique: "add_goal"}, h.AddGoalHandler)
	bot.Handle(&tele.Btn{Unique: "priority_low"}, h.SetPriorityHandler)
	bot.Handle(&tele.Btn{Unique: "priority_medium"}, h.SetPriorityHandler)
	bot.Handle(&tele.Btn{Unique: "priority_high"}, h.SetPriorityHandler)

	bot.Handle(&tele.Btn{Unique: "statistic_panel"}, h.onStatisticPanel)
	bot.Handle(&tele.Btn{Unique: "statistic"}, h.onStatistic)
	bot.Handle(&tele.Btn{Unique: "main_menu"}, h.onMainMenu)
	bot.Handle(&tele.Btn{Unique: "back_to_panel"}, h.onBackToPanel)

	bot.Handle(&tele.Btn{Unique: "goals_panel"}, h.onGoalsPanel)
	bot.Handle(&tele.Btn{Unique: "add_goal"}, h.AddGoalHandler)
	bot.Handle(&tele.Btn{Unique: "list_goal"}, h.onListGoal)
	bot.Handle(&tele.Btn{Unique: "edit_goal"}, h.onEditGoal)
	bot.Handle(&tele.Btn{Unique: "edit_goal_select"}, h.onEditGoalSelect)
	bot.Handle(&tele.Btn{Unique: "edit_field"}, h.onEditField)
	bot.Handle(&tele.Btn{Unique: "edit_priority_select"}, h.onEditPrioritySelect)
	bot.Handle(&tele.Btn{Unique: "edit_status_select"}, h.onEditStatusSelect)
	bot.Handle(&tele.Btn{Unique: "delete_goal"}, h.onDeleteGoal)
	bot.Handle(&tele.Btn{Unique: "delete_goal_confirm"}, h.onDeleteGoalConfirm)

	// Управление платежами

	bot.Handle(&tele.Btn{Unique: "sber"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "vtb"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "sbp"}, h.onBankDetails)

	// Кнопки возврата
	bot.Handle(&tele.Btn{Unique: "back"}, h.onBack)
	bot.Handle(&tele.Btn{Unique: "backAlbum"}, h.onBackAlbum)

}

func (h *Handler) getOrCreateFSM(userID int64) *fsm.FSM {
	if _, exists := h.UserFSM[userID]; !exists {
		h.UserFSM[userID] = fsm.NewFSM(
			StateStart,
			fsm.Events{
				{Name: "bank", Src: []string{StateStart}, Dst: StateSelectBank},
				{Name: "amount", Src: []string{StateSelectBank}, Dst: StateEnterAmount},
				{Name: "finish", Src: []string{StateEnterAmount}, Dst: StateFinish},
			},
			fsm.Callbacks{},
		)
	}
	return h.UserFSM[userID]
}

func (h *Handler) getOrCreateAdminFSM(userID int64) *fsm.FSM {
	if _, exists := h.AdminFSM[userID]; !exists {
		// Добавляем состояния и переходы для редактирования
		h.AdminFSM[userID] = fsm.NewFSM(
			StateStart,
			fsm.Events{
				// Добавление новой цели
				{Name: "add_goal_title", Src: []string{StateStart}, Dst: StateAddGoalTitle},
				{Name: "add_goal_description", Src: []string{StateAddGoalTitle}, Dst: StateAddGoalDescription},
				{Name: "add_goal_target_sum", Src: []string{StateAddGoalDescription}, Dst: StateAddGoalTargetSum},
				{Name: "finish_goal", Src: []string{StateAddGoalTargetSum}, Dst: StateFinishedGoal},

				// Редактирование цели
				{Name: "go_edit_goal_select", Src: []string{StateStart}, Dst: StateEditGoalSelect},
				{Name: "go_edit_goal_field", Src: []string{StateEditGoalSelect}, Dst: StateEditGoalFieldSelect},
				{Name: "wait_input", Src: []string{StateEditGoalFieldSelect}, Dst: StateEditGoalWaitInput},
				// Можно добавить другие переходы для завершения редактирования, если нужно
			},
			fsm.Callbacks{},
		)
	}
	return h.AdminFSM[userID]
}

// Обработчик текстовых сообщений с учетом режима пользователя
func (h *Handler) onText(c tele.Context) error {
	mode, _ := h.UserData[c.Sender().ID]["mode"].(string)

	// Если пользователь в админском режиме
	if mode == "admin" {
		if adminFSM, exists := h.AdminFSM[c.Sender().ID]; exists {
			switch adminFSM.Current() {
			case StateAddGoalTitle:
				return h.processGoalTitle(c, adminFSM)
			case StateAddGoalDescription:
				return h.processGoalDescription(c, adminFSM)
			case StateAddGoalTargetSum:
				return h.processGoalTargetSum(c, adminFSM)
			case StateEditGoalWaitInput:
				// Здесь обрабатываем текстовый ввод для редактирования цели
				return h.onTextAdminEdit(c, adminFSM)
			default:
				log.Printf("Неизвестное состояние FSM администратора у пользователя %d: %s", c.Sender().ID, adminFSM.Current())
				h.resetFSM(c.Sender().ID)
				return c.Send("Неверный ввод. Начните процесс заново или используйте /panel для возвращения в админ-панель.")
			}
		} else {
			h.resetFSM(c.Sender().ID)
			return c.Send("Нет активного админского процесса. Попробуйте снова.")
		}
	}

	// Если пользователь в пользовательском режиме (пожертвования)
	if mode == "user" {
		if userFSM, exists := h.UserFSM[c.Sender().ID]; exists {
			switch userFSM.Current() {
			case StateSelectBank:
				return h.onBankDetails(c)
			case StateEnterAmount:
				return h.onEnterAmount(c)
			default:
				log.Printf("Неизвестное состояние FSM пожертвований у пользователя %d: %s", c.Sender().ID, userFSM.Current())
				h.resetFSM(c.Sender().ID)
				return c.Send("Неверный ввод. Начните процесс заново.")
			}
		} else {
			h.resetFSM(c.Sender().ID)
			return c.Send("Нет активного процесса пожертвования. Введите /start для начала.")
		}
	}

	// Если режим не установлен
	h.resetFSM(c.Sender().ID)
	return c.Send("У вас нет активного процесса. Введите /start, чтобы начать.")
}
