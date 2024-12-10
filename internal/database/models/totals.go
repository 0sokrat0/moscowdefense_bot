package models

import "time"

type TotalDonation struct {
	ID        int       `gorm:"primaryKey;autoIncrement"` // Основной ключ
	Total     float64   `gorm:"not null"`                 // Общая сумма всех пожертвований
	UpdatedAt time.Time `gorm:"autoUpdateTime"`           // Последнее обновление суммы
}
