package handlers

import (
	tele "gopkg.in/telebot.v4"
	"log"
)

// onInfo обрабатывает кнопку "Информация о фонде"
func (h *Handler) onInfo(c tele.Context) error {

	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	menu := &tele.ReplyMarkup{}
	btnBack := menu.Data("⬅️ Назад", "back") // Кнопка для возврата

	menu.Inline(
		menu.Row(btnBack), // Добавляем кнопку в меню
	)

	return c.Send(
		"<b>О нас:</b>\n«Марфинский Тыл» поддерживает бойцов 108-го гв. дшп. Наша миссия — стать надежным тылом и помощью для героев.\n<b>Присоединяйтесь к нашей инициативе!</b>",
		menu,
	)
}
