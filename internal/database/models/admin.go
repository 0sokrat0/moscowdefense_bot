package models

type Admin struct {
	TgID     int64  `gorm:"unique;not null"` // Telegram ID администратора
	Username string `gorm:"unique"`          // Имя пользователя Telegram
	Role     string `gorm:"default:user"`    // Роль (например, "user", "admin", "superadmin")
}
