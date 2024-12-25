package handlers

import (
	tele "gopkg.in/telebot.v4"
)

// Создание кнопки "Назад"
func backButton(data string) *tele.ReplyMarkup {
	back := &tele.ReplyMarkup{}
	BackBtn := back.Data("⬅️ Назад", data)
	back.Inline(back.Row(BackBtn))
	return back
}

// Админ-панель (главное меню)
func (h *Handler) onPanel(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к админ-панели.")
	}

	menu := &tele.ReplyMarkup{}
	GoalsBtn := menu.Data("📝 Управление целями", "goals_panel")
	StatisticBtn := menu.Data("📊 Статистика", "statistic_panel")
	BroadcastBtn := menu.Data("📨 Рассылка", "broadcast_panel")
	AddAdminBtn := menu.Data("➕ Добавить администратора", "add_admin")
	BalanceBtn := menu.Data("💰 Управление балансом", "balance_panel")
	BackBtn := menu.Data("🔙 Назад", "back")

	menu.Inline(
		menu.Row(GoalsBtn, StatisticBtn),
		menu.Row(BroadcastBtn),
		menu.Row(AddAdminBtn),
		menu.Row(BalanceBtn),
		menu.Row(BackBtn),
	)

	return c.Send("Админ панель", menu)
}

func (h *Handler) onBackToPanel(c tele.Context) error {
	h.tryDeleteMessage(c)
	return h.onPanel(c)
}
