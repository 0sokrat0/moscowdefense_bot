package handlers

import (
	"TgDonation/internal/database/models"
	"log"

	tele "gopkg.in/telebot.v4"
	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:                db,
		UserAlbumMessages: make(map[int64][]*tele.Message),
	}
}

func (h *Handler) onInfo(c tele.Context) error {
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	menu := &tele.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", "backAlbum")
	menu.Inline(menu.Row(btnBack))

	photoPaths := []string{
		"pkg/img/photo_1.jpg",
		"pkg/img/photo_2.jpg",
		"pkg/img/photo_3.jpg",
		"pkg/img/photo_4.jpg",
		"pkg/img/photo_10.jpg",
	}

	fileIDs, err := h.uploadPhotos(c, photoPaths)
	if err != nil {
		log.Printf("Ошибка загрузки фотографий: %v", err)
		return err
	}

	if err := h.sendAlbumByIDs(c, fileIDs); err != nil {
		log.Printf("Ошибка отправки альбома: %v", err)
		return err
	}

	text := `<b>О нас:</b>
«Марфинский Тыл» поддерживает бойцов из 108-й гвардейский десантно-штурмовой Кубанский казачий ордена Красной Звезды полк (108 гв. дшп).
Для бойцов мы являемся надежным тылом и опорой. Прошу всех неравнодушных присоединяться и помогать в одно ногу ✊
Слава России 🇷🇺

Наша миссия — стать надежным тылом и опорой для героев.
<b>Присоединяйтесь к нашей инициативе!</b>`

	if err := c.Send(text, menu); err != nil {
		log.Printf("Ошибка отправки текста: %v", err)
		return err
	}

	return nil
}

func (h *Handler) uploadPhotos(c tele.Context, photoPaths []string) ([]string, error) {
	fileIDs := []string{}

	for _, path := range photoPaths {
		fileID, found, err := models.GetFileID(h.DB, path)
		if err != nil {
			return nil, err
		}

		if found {
			fileIDs = append(fileIDs, fileID)
			continue
		}

		photo := &tele.Photo{File: tele.FromDisk(path)}
		msg, err := c.Bot().Send(c.Chat(), photo)
		if err != nil {
			log.Printf("Ошибка загрузки фото %s: %v", path, err)
			return nil, err
		}

		if err := models.SaveFileID(h.DB, path, msg.Photo.FileID); err != nil {
			log.Printf("Ошибка сохранения file_id для фото %s: %v", path, err)
			return nil, err
		}

		fileIDs = append(fileIDs, msg.Photo.FileID)
	}

	return fileIDs, nil
}

func (h *Handler) sendAlbumByIDs(c tele.Context, fileIDs []string) error {
	album := tele.Album{}

	for _, fileID := range fileIDs {
		album = append(album, &tele.Photo{File: tele.File{FileID: fileID}})
	}

	msgs, err := c.Bot().SendAlbum(c.Chat(), album)
	if err != nil {
		log.Printf("Ошибка отправки альбома: %v", err)
		return err
	}

	msgPtrs := make([]*tele.Message, len(msgs))
	for i := range msgs {
		msgPtrs[i] = &msgs[i]
	}

	if h.UserAlbumMessages == nil {
		h.UserAlbumMessages = make(map[int64][]*tele.Message)
	}

	h.UserAlbumMessages[c.Sender().ID] = msgPtrs

	return nil
}

func (h *Handler) onBackAlbum(c tele.Context) error {
	// Удаляем текущее сообщение
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	// Получаем ID пользователя
	userID := c.Sender().ID

	// Удаляем все сообщения альбома для данного пользователя
	if msgs, ok := h.UserAlbumMessages[userID]; ok {
		ch := make(chan error, len(msgs)) // Канал для сбора ошибок
		for _, msg := range msgs {
			go func(m *tele.Message) {
				if err := c.Bot().Delete(m); err != nil {
					log.Printf("Ошибка удаления сообщения альбома: %v", err)
					ch <- err
				} else {
					ch <- nil
				}
			}(msg)
		}

		// Проверяем результаты удаления
		for range msgs {
			if err := <-ch; err != nil {
				log.Printf("Ошибка при удалении одного из сообщений альбома: %v", err)
			}
		}

		// Удаляем записи о сообщениях альбома
		delete(h.UserAlbumMessages, userID)
	}

	// Возвращаемся к стартовому меню
	return h.onStart(c)
}
