package models

import "time"

type User struct {
	ID          int        `gorm:"primaryKey;autoIncrement"` // Основной ключ
	TgID        int64      `gorm:"unique;not null"`          // Уникальный Telegram ID
	Username    string     `gorm:"unique"`                   // Уникальное имя пользователя
	PhoneNumber string     `gorm:"unique"`                   // Уникальный номер телефона (опционально)
	CreatedAt   time.Time  `gorm:"autoCreateTime"`           // Дата регистрации
	Donations   []Donation `gorm:"foreignKey:UserID"`        // Связь с таблицей пожертвований
}
