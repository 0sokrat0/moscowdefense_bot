package handlers

import (
	"github.com/looplab/fsm"
	tele "gopkg.in/telebot.v4"
	"gorm.io/gorm"
)

const (
	StateDonation = "donation"
	StateFinished = "finished"
)

type Handler struct {
	Bot               *tele.Bot
	DB                *gorm.DB
	UserFSM           map[int64]*fsm.FSM
	UserData          map[int64]map[string]interface{}
	UserAlbumMessages map[int64][]*tele.Message
}

func RegisterHandlers(bot *tele.Bot, db *gorm.DB, userFSM map[int64]*fsm.FSM, userData map[int64]map[string]interface{}) {
	h := &Handler{
		Bot:      bot,
		DB:       db,
		UserFSM:  userFSM,
		UserData: userData,
	}

	// Команды
	bot.Handle("/start", h.onStart)
	bot.Handle(tele.OnContact, h.onContact)

	// Кнопки
	menu := &tele.ReplyMarkup{}
	btnInfo := menu.Data("ℹ️ Информация о фонде", "info")
	btnSocial := menu.Data("💬 Наши соц.сети", "social")
	btnGoal := menu.Data("🎯 Цели", "goal")

	bot.Handle(&tele.Btn{Unique: "donation"}, h.onDonation)
	bot.Handle(&tele.Btn{Unique: "sber"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "vtb"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "sbp"}, h.onBankDetails)
	bot.Handle(&tele.Btn{Unique: "back"}, h.onBack)
	bot.Handle(&tele.Btn{Unique: "backAlbum"}, h.onBackAlbum)
	bot.Handle(&btnInfo, h.onInfo)
	bot.Handle(&tele.Btn{Unique: "main_menu"}, h.onMainMenu)
	bot.Handle(tele.OnText, h.onEnterAmount)
	bot.Handle(tele.OnPhoto, h.onUploadReceipt)
	bot.Handle(&btnSocial, h.onSocial)
	bot.Handle(&btnGoal, h.onGoal)

}

func (h *Handler) getOrCreateFSM(userID int64) *fsm.FSM {
	if _, exists := h.UserFSM[userID]; !exists {
		h.UserFSM[userID] = fsm.NewFSM(
			"start",
			fsm.Events{
				{Name: "bank", Src: []string{"start"}, Dst: StateSelectBank},
				{Name: "amount", Src: []string{StateSelectBank}, Dst: StateEnterAmount},
				{Name: "finish", Src: []string{StateEnterAmount}, Dst: StateFinish},
			},
			fsm.Callbacks{},
		)
	}
	return h.UserFSM[userID]
}
