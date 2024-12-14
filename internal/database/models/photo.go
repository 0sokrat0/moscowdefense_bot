package models

import (
	"gorm.io/gorm"
)

// Photo представляет модель для хранения пути и file_id.
type Photo struct {
	ID     uint   `gorm:"primaryKey"`
	Path   string `gorm:"unique;not null"` // Локальный путь к фотографии
	FileID string `gorm:"unique;not null"` // Telegram file_id
}

// GetFileID возвращает file_id для указанного пути, если он существует.
func GetFileID(db *gorm.DB, path string) (string, bool, error) {
	var photo Photo
	result := db.Where("path = ?", path).First(&photo)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, result.Error
	}
	return photo.FileID, true, nil
}

// SaveFileID сохраняет file_id для указанного пути.
func SaveFileID(db *gorm.DB, path string, fileID string) error {
	photo := Photo{
		Path:   path,
		FileID: fileID,
	}
	return db.Create(&photo).Error
}
