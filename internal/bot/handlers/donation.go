package handlers

import (
	"TgDonation/internal/database/models"
	"context"
	tele "gopkg.in/telebot.v4"
	"log"
	"strconv"
)

const (
	StateStart       = "start"
	StateSelectBank  = "bank"
	StateEnterAmount = "amount"
	StateFinish      = "finish"
)

func (h *Handler) onDonation(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	if err := fsm.Event(ctx, "bank"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка. Начните процесс заново, выбрав 'Сделать пожертвование'.")
	}

	menu := &tele.ReplyMarkup{}
	btnSber := menu.Data("🟢 Сбербанк", "sber", "sber")
	btnVTB := menu.Data("🔵 ВТБ", "vtb", "vtb")
	btnSBP := menu.Data("💠 СБП", "sbp", "sbp")

	menu.Inline(
		menu.Row(btnSber, btnVTB, btnSBP),
	)

	return c.Send("<b>Выберите банк для пожертвования:</b>", menu)
}

func (h *Handler) onBankDetails(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	log.Printf("FSM Current State (Before Event) for User %d: %s", c.Sender().ID, fsm.Current())

	if fsm.Current() != StateSelectBank {
		h.resetFSM(c.Sender().ID)
		return c.Send("Выбор банка сейчас недоступен. Начните процесс заново.")
	}

	if h.UserData[c.Sender().ID] == nil {
		h.UserData[c.Sender().ID] = make(map[string]interface{})
	}

	selectedBank := c.Callback().Data
	var bankDetails string
	switch selectedBank {
	case "sber":
		h.UserData[c.Sender().ID]["bank"] = "Сбербанк"
		bankDetails = "🟢 <b>Реквизиты Сбербанка:</b>\nКарта:<code> 2202 2080 3701 1005</code>\n<b>Получатель:</b> Лукичёва Юлия Николаевна"
	case "vtb":
		h.UserData[c.Sender().ID]["bank"] = "ВТБ"
		bankDetails = "🔵 <b>Реквизиты ВТБ:</b>\nКарта:<code> 2200 2402 1368 9108</code>\n<b>Получатель:</b> Лукичёва Юлия Николаевна"
	case "sbp":
		h.UserData[c.Sender().ID]["bank"] = "СБП"
		bankDetails = "💠 <b>Реквизиты СБП:</b>\nТелефон:<code> +7 963 752-92-99</code>\n<b>Получатель:</b> Лукичёва Юлия Николаевна"
	default:
		return c.Send("Неизвестный банк. Попробуйте снова.")
	}

	if err := fsm.Event(ctx, "amount"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка. Начните процесс заново.")
	}

	log.Printf("User %d selected bank: %s", c.Sender().ID, selectedBank)

	return c.Send(bankDetails + "\n\n<b>Введите сумму пожертвования:</b>")
}

func (h *Handler) onEnterAmount(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)

	log.Printf("FSM Current State for User %d: %s", c.Sender().ID, fsm.Current())

	if fsm.Current() != StateEnterAmount {
		h.resetFSM(c.Sender().ID)
		return c.Send("Вы не можете ввести сумму сейчас. Начните процесс заново.")
	}

	if h.UserData[c.Sender().ID] == nil {
		h.UserData[c.Sender().ID] = make(map[string]interface{})
	}
	h.UserData[c.Sender().ID]["amount"] = c.Text()

	log.Printf("User %d entered amount: %s", c.Sender().ID, c.Text())

	return c.Send("<b>Пожалуйста, загрузите фото чека для подтверждения:</b>")
}

func (h *Handler) onUploadReceipt(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	back := &tele.ReplyMarkup{}
	btnBack := back.Data("⬅️ Назад", "back")

	back.Inline(
		back.Row(btnBack),
	)

	log.Printf("FSM Current State for User %d: %s", c.Sender().ID, fsm.Current())

	if fsm.Current() != StateEnterAmount {
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Вы не можете загрузить чек сейчас. Начните процесс заново.</b>")
	}

	if c.Message().Photo == nil {
		return c.Send("<b>Пожалуйста, отправьте фото чека.</b>")
	}

	// Получение последнего фото из массива (Telegram отправляет фото с разными размерами)
	photo := c.Message().Photo
	fileID := photo.FileID

	var user models.User
	if err := h.DB.FirstOrCreate(&user, models.User{TgID: c.Sender().ID}).Error; err != nil {
		log.Printf("Database Error (User): %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Ошибка при обработке пользователя. Попробуйте позже.</b>")
	}

	// Получаем данные из UserData
	bank, bankExists := h.UserData[c.Sender().ID]["bank"]
	if !bankExists {
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Не удалось найти данные о банке. Начните процесс заново.</b>")
	}

	amountStr, amountExists := h.UserData[c.Sender().ID]["amount"].(string)
	if !amountExists {
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Не удалось найти данные о сумме. Начните процесс заново.</b>")
	}

	// Преобразование суммы в число
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Сумма введена некорректно. Начните процесс заново.</b>")
	}

	// Сохраняем пожертвование в базе данных
	donation := models.Donation{
		UserID:       user.ID,
		BankName:     bank.(string),
		Amount:       float64(amount),
		ReceiptPhoto: fileID, // Сохраняем FileID фото
	}

	if err := h.DB.Create(&donation).Error; err != nil {
		log.Printf("Database Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Ошибка при сохранении данных пожертвования. Попробуйте позже.</b>")
	}

	// Завершаем FSM
	if err := fsm.Event(ctx, "finish"); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Произошла ошибка. Начните процесс заново.</b>")
	}

	// Удаляем данные пользователя и FSM
	delete(h.UserData, c.Sender().ID)
	delete(h.UserFSM, c.Sender().ID)

	return c.Send("<b>Спасибо за ваше пожертвование! Ваша поддержка важна.</b>", back)
}

func (h *Handler) resetFSM(userID int64) {
	log.Printf("Resetting FSM for User %d", userID)
	if h.UserFSM[userID] != nil {
		h.UserFSM[userID].SetState(StateStart)
	}
	delete(h.UserData, userID)
}
