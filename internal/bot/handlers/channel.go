package handlers

import tele "gopkg.in/telebot.v4"

// onDonation обрабатывает кнопку "Сделать пожертвование"
func (h *Handler) onChannel(c tele.Context) error {
	return c.Send("Спасибо за ваше желание помочь! Реквизиты для пожертвований: ...")
}
