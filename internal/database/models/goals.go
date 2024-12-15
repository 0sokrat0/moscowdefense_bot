package models

import "time"

type Goal struct {
	ID           uint   `gorm:"primaryKey"`
	GoalText     string `gorm:"type:text;not null"`
	Title        string `gorm:"type:text;not null"`
	Description  string
	TargetSum    float64 `gorm:"not null"`
	CurrentSum   float64 `gorm:"default:0"`
	AllocatedSum float64 `gorm:"default:0"` // Зарезервированная сумма
	Status       string  `gorm:"default:'active'"`
	Priority     string  `gorm:"default:'medium'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AdminID      uint       `gorm:"not null"`
	DeletedAt    *time.Time `gorm:"index"`
}
