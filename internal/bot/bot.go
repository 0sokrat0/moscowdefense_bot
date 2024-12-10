package bot

import (
	"TgDonation"
	"TgDonation/internal/bot/handlers"

	"log"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"

	"gorm.io/gorm"
)

type Bot struct {
	*tele.Bot
	db *gorm.DB
}

func New(token string, boot TgDonation.Bootstrap) (*Bot, error) {
	// Проверка токена перед вызовом NewBot
	log.Printf("Инициализация бота с токеном: %s", token)

	b, err := tele.NewBot(tele.Settings{
		Token:     token,
		ParseMode: "HTML",
		Poller:    &tele.LongPoller{Timeout: 5 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	return &Bot{
		Bot: b,
		db:  boot.DB,
	}, nil
}

func (b *Bot) Start() {
	b.Use(middleware.Logger())
	b.Use(middleware.AutoRespond())

	handlers.RegisterHandlers(b.Bot, b.db)

	log.Println("Бот успешно запущен")
	b.Bot.Start()
}
