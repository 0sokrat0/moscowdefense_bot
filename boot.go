package TgDonation

import "gorm.io/gorm"

// Bootstrap содержит глобальные зависимости приложения
type Bootstrap struct {
	DB *gorm.DB // Подключение к базе данных через GORM
	// Можно добавить другие зависимости, такие как клиент для API, кэш, логгер и т.д.
}
