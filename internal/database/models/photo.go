package models

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
			log.Printf("Файл с путем %s не найден", path)
			return "", false, nil
		}
		log.Printf("Ошибка при поиске файла %s: %v", path, result.Error)
		return "", false, result.Error
	}
	return photo.FileID, true, nil
}

// SaveFileID сохраняет или обновляет file_id для указанного пути.
func SaveFileID(db *gorm.DB, path string, fileID string) error {
	photo := Photo{
		Path:   path,
		FileID: fileID,
	}

	// Используем INSERT ON DUPLICATE KEY UPDATE (GORM clause.Upsert)
	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}},               // Конфликт по полю Path
		DoUpdates: clause.AssignmentColumns([]string{"file_id"}), // Обновляем file_id
	}).Create(&photo).Error

	if err != nil {
		log.Printf("Ошибка сохранения file_id для %s: %v", path, err)
		return err
	}

	return nil
}
