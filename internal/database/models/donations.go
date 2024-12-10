package models

import "time"

type Donation struct {
	ID        int       `gorm:"primaryKey;autoIncrement"` // Основной ключ
	UserID    int       `gorm:"not null"`                 // Внешний ключ к таблице пользователей
	Amount    float64   `gorm:"not null"`                 // Сумма пожертвования
	CreatedAt time.Time `gorm:"autoCreateTime"`           // Дата пожертвования
}