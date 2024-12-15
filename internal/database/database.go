package database

import (
	"TgDonation/internal/database/models"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func DBConnect(databasePath string) (*gorm.DB, error) {
	// Подключение к SQLite
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Автоматическая миграция
	err = db.AutoMigrate(&models.User{}, &models.Donation{}, &models.TotalDonation{}, &models.Goal{}, &models.Photo{}, &models.Admin{})
	if err != nil {
		return nil, err
	}

	// Проверяем, что подключение не равно nil
	if db == nil {
		return nil, fmt.Errorf("gorm.Open вернул nil экземпляр базы данных")
	}

	return db, nil
}
