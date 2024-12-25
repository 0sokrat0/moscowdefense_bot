package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"TgDonation/internal/database/models"
	"github.com/looplab/fsm"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onBalancePanel(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	menu := &tele.ReplyMarkup{}
	AddFunds := menu.Data("➕ Добавить средства", "add_funds")
	SubFunds := menu.Data("➖ Вычесть средства", "sub_funds")
	BackBtn := menu.Data("⬅️ Назад", "back_to_panel")

	menu.Inline(
		menu.Row(AddFunds),
		menu.Row(SubFunds),
		menu.Row(BackBtn),
	)

	return c.Send("Управление общим балансом:", menu)
}

func (h *Handler) onAddFundsStart(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsmObj := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"action": "add_funds",
		"mode":   "admin",
	}

	if err := fsmObj.Event(context.Background(), "wait_add_funds_amount"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Попробуйте снова.")
	}

	return c.Send("Введите сумму для добавления к общему балансу:")
}

func (h *Handler) processAddFunds(c tele.Context, fsmObj *fsm.FSM) error {
	amountStr := c.Text()
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Send("Некорректная сумма. Введите положительное число.")
	}

	if err := h.addToTotalDonation(amount); err != nil {
		log.Printf("Ошибка при добавлении средств: %v", err)
		return c.Send("Ошибка при добавлении средств.")
	}

	h.resetFSM(c.Sender().ID)
	delete(h.UserData, c.Sender().ID)
	return c.Send(fmt.Sprintf("✅ Добавлено %.2f к общему балансу", amount), backButton("back_to_panel"))
}

func (h *Handler) onSubFundsStart(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsmObj := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"action": "sub_funds",
		"mode":   "admin",
	}

	if err := fsmObj.Event(context.Background(), "wait_sub_funds_amount"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Попробуйте снова.")
	}

	return c.Send("Введите сумму для вычета из общего баланса:")
}

func (h *Handler) processSubFunds(c tele.Context, fsmObj *fsm.FSM) error {
	amountStr := c.Text()
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Send("Некорректная сумма. Введите положительное число.")
	}

	if err := h.subtractFromTotalDonation(amount); err != nil {
		log.Printf("Ошибка при вычитании средств: %v", err)
		return c.Send("Ошибка при вычитании средств.")
	}

	h.resetFSM(c.Sender().ID)
	delete(h.UserData, c.Sender().ID)
	return c.Send(fmt.Sprintf("✅ Вычтено %.2f из общего баланса", amount), backButton("back_to_panel"))
}

// Методы работы с моделью TotalDonation
func (h *Handler) addToTotalDonation(amount float64) error {
	var totalRec models.TotalDonation
	err := h.DB.First(&totalRec).Error
	if err != nil {
		// Если записи нет, создаём
		totalRec.Total = amount
		return h.DB.Create(&totalRec).Error
	}
	totalRec.Total += amount
	return h.DB.Save(&totalRec).Error
}

func (h *Handler) subtractFromTotalDonation(amount float64) error {
	var totalRec models.TotalDonation
	err := h.DB.First(&totalRec).Error
	if err != nil {
		// Нет записей или ошибка – тогда нечего вычитать
		return fmt.Errorf("нет доступных средств для вычета")
	}
	if totalRec.Total < amount {
		// Если пытаемся вычесть больше, чем есть, уменьшим до нуля
		amount = totalRec.Total
	}
	totalRec.Total -= amount
	return h.DB.Save(&totalRec).Error
}
