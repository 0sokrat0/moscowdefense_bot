package handlers

import tele "gopkg.in/telebot.v4"

func (h *Handler) onBroadcast(c tele.Context) error {
	h.tryDeleteMessage(c)
	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", "back_to_panel")
	back.Inline(
		back.Row(BackBtn),
	)
	return c.Send("Рассылка пока в разработке...", back)
}
