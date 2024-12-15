package models

import "time"

type Goal struct {
	ID          uint   `gorm:"primaryKey"`
	GoalText    string `gorm:"type:text;not null"` // Исправлено: соответствует goal_text в базе данных
	Title       string `gorm:"type:text;not null"`
	Description string
	TargetSum   float64    `gorm:"not null"`
	CurrentSum  float64    `gorm:"default:0"`
	Status      string     `gorm:"type:text;default:'active'"`
	Priority    string     `gorm:"type:text;default:'medium'"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	AdminID     uint       `gorm:"not null"`
	DeletedAt   *time.Time `gorm:"index"`
}
