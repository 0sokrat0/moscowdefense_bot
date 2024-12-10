package handlers

import tele "gopkg.in/telebot.v4"

// onInfo обрабатывает кнопку "Информация о фонде"
func (h *Handler) onInfo(c tele.Context) error {
	return c.Send("Мы - Марфинский Тыл, поддерживаем тех, кто участвует в СВО. Подробности на нашем сайте.")
}
