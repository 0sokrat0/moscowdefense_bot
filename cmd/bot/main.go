package main

import (
	"log"

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

	// Инициализация бота
	b, err := bot.New(config.Token, boot)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	// Запуск бота
	log.Println("Бот успешно запущен")
	b.Start()
}
