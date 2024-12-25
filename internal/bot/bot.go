package bot

import (
	"TgDonation"
	"TgDonation/internal/bot/handlers"

	"github.com/looplab/fsm"

	"log"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"

	"gorm.io/gorm"
)

type Bot struct {
	*tele.Bot
	db          *gorm.DB
	userFSM     map[int64]*fsm.FSM               // FSM для каждого пользователя
	userData    map[int64]map[string]interface{} // Данные пользователей
	GroupChatID int64
}

func New(token string, boot TgDonation.Bootstrap, groupChatID int64) (*Bot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:     token,
		ParseMode: "HTML",
		Poller:    &tele.LongPoller{Timeout: 5 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	return &Bot{
		Bot:         b,
		db:          boot.DB,
		userFSM:     make(map[int64]*fsm.FSM),
		userData:    make(map[int64]map[string]interface{}),
		GroupChatID: groupChatID,
	}, nil
}

func (b *Bot) Start() {
	// b.Use(middleware.Logger())
	b.Use(middleware.AutoRespond())

	handlers.RegisterHandlers(b.Bot, b.db, b.userFSM, b.userData, b.GroupChatID)

	log.Println("Бот успешно запущен")
	b.Bot.Start()
}
