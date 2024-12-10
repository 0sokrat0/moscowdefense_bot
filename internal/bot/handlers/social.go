package handlers

import tele "gopkg.in/telebot.v4"

// onDonation обрабатывает кнопку "Сделать пожертвование"
func (h *Handler) onSocial(c tele.Context) error {
	menu := &tele.ReplyMarkup{}
	btn1 := menu.URL("1️⃣ Канал", "https://t.me/+2paESwUQmWdlOTU6")
	btn2 := menu.URL("2️⃣ Чат", "https://t.me/+F2bcWe3FVg9jODEy")
	btn3 := menu.URL("3️⃣ Связь с организатором", "https://t.me/ligr_91")

	menu.Inline(
		menu.Row(btn1),
		menu.Row(btn2),
		menu.Row(btn3),
	)

	c.Send("<b>Если у вас есть вопросы, предложения или вы хотите предложить помощь, свяжитесь с нами:</b>", menu)

	return nil
}
