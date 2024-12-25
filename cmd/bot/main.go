package main

import (
	"log"
	"strconv"

	"TgDonation"
	"TgDonation/internal/bot"
	"TgDonation/internal/configs"
	"TgDonation/internal/database"
)

func main() {
	// Загрузка конфигурации с использованием Viper
	config, err := configs.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключение к базе данных
	db, err := database.DBConnect(config.DBPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Инициализация зависимостей через Bootstrap
	boot := TgDonation.Bootstrap{
		DB: db,
	}

	groupChatID, err := strconv.ParseInt(config.GroupChatID, 10, 64)
	if err != nil {
		log.Fatalf("Невалидный GROUP_CHAT_ID: %v", err)
	}

	// Инициализация бота
	b, err := bot.New(config.Token, boot, groupChatID)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	b.Start()
}
