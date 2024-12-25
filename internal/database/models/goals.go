package models

import "time"

type Goal struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"not null"`
	Description string
	TargetSum   float64 `gorm:"not null"`
	Status      string  `gorm:"not null;default:active"`
	Priority    string
	AdminID     uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Возможные статусы (пример):
// - "active"
// - "inactive"
// - "finished"
// - "canceled"
