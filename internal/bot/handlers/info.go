package handlers

import (
	"log"

	tele "gopkg.in/telebot.v4"
)

// onInfo обрабатывает кнопку "Информация о фонде"
func (h *Handler) onInfo(c tele.Context) error {

	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	menu := &tele.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", "back")

	menu.Inline(
		menu.Row(btnBack),
	)

	return c.Send(
		"<b>О нас:</b><br>«Марфинский Тыл» поддерживает бойцов из 108-й гвардейский десантно-штурмовой Кубанский казачий ордена Красной Звезды полк (108 гв. дшп).<br>Для бойцов мы являемся надежным тылом и опорой. Прошу всех неравнодушных присоединяться и помогать в одно ногу ✊<br>Слава России 🇷🇺<br><br>Наша миссия — стать надежным тылом и опорой для героев.<br><b>Присоединяйтесь к нашей инициативе!</b>",
		menu,
	)
}
