package models

import "time"

type Donation struct {
	ID           int     `gorm:"primaryKey;autoIncrement"` // Основной ключ
	UserID       int     `gorm:"not null"`                 // Внешний ключ к таблице пользователей
	BankName     string  `gorm:"not null"`                 // Имя банка
	User         User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Amount       float64 `gorm:"not null"` // Сумма пожертвования
	ReceiptPhoto string
	CreatedAt    time.Time `gorm:"autoCreateTime"` // Дата пожертвования
}
