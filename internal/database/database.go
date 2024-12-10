package database

import (
	"TgDonation/internal/database/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBConnect устанавливает соединение с базой данных SQLite и выполняет миграции
func DBConnect(databasePath string) (*gorm.DB, error) {
	// Подключение к базе данных SQLite
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Автоматическая миграция схемы
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
