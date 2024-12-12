package models

import "time"

type Goal struct {
	ID          uint       `gorm:"primaryKey"`                 // Уникальный идентификатор цели
	GoalText    string     `gorm:"type:text;not null"`         // Текст описания цели
	Status      string     `gorm:"type:text;default:'active'"` // Статус цели (active, completed, cancelled)
	Priority    string     `gorm:"type:text;default:'medium'"` // Приоритет цели (low, medium, high)
	CreatedAt   time.Time  `gorm:"autoCreateTime"`             // Время создания
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`             // Время обновления
	CompletedAt *time.Time `gorm:"default:null"`               // Время завершения
	AdminID     uint       `gorm:"not null"`                   // ID администратора
	DeletedAt   *time.Time `gorm:"index"`                      // Поле для мягкого удаления
}
