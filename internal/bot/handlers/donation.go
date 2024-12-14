package handlers

import (
	"TgDonation/internal/database/models"
	"context"
	"log"
	"strconv"

	tele "gopkg.in/telebot.v4"
)

const (
	StateStart       = "start"
	StateSelectBank  = "bank"
	StateEnterAmount = "amount"
	StateFinish      = "finish"
)

type FSM interface {
	Event(ctx context.Context, event string) error
	Current() string
	SetState(state string)
}

func (h *Handler) onDonation(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	if err := deleteMessage(c); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	if err := fsm.Event(ctx, StateSelectBank); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка. Начните процесс заново, выбрав 'Сделать пожертвование'.")
	}

	menu := createBankMenu()
	return c.Send("<b>Выберите банк для пожертвования:</b>", menu)
}

func (h *Handler) onBankDetails(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	if err := deleteMessage(c); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	if fsm.Current() != StateSelectBank {
		h.resetFSM(c.Sender().ID)
		return c.Send("Выбор банка сейчас недоступен. Начните процесс заново.")
	}

	h.UserData[c.Sender().ID] = ensureUserData(h.UserData[c.Sender().ID])
	selectedBank := c.Callback().Data
	bankDetails, valid := getBankDetails(selectedBank)
	if !valid {
		return c.Send("Неизвестный банк. Попробуйте снова.")
	}

	h.UserData[c.Sender().ID]["bank"] = bankDetails.BankName
	if err := fsm.Event(ctx, StateEnterAmount); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка. Начните процесс заново.")
	}

	log.Printf("User %d selected bank: %s", c.Sender().ID, selectedBank)
	return c.Send(bankDetails.Details + "\n\n<b>Введите сумму пожертвования:</b>")
}

func (h *Handler) onEnterAmount(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	if fsm.Current() != StateEnterAmount {
		h.resetFSM(c.Sender().ID)
		return c.Send("Вы не можете ввести сумму сейчас. Начните процесс заново.")
	}

	reaction := tele.Reaction{
		Type:  "emoji",
		Emoji: "👌",
	}

	reactions := tele.Reactions{
		Reactions: []tele.Reaction{reaction},
		Big:       false,
	}

	if err := c.Bot().React(c.Sender(), c.Message(), reactions); err != nil {
		log.Printf("Ошибка при добавлении реакции: %v", err)
		return c.Send("Не удалось добавить реакцию.")
	}

	h.UserData[c.Sender().ID] = ensureUserData(h.UserData[c.Sender().ID])
	h.UserData[c.Sender().ID]["amount"] = c.Text()
	log.Printf("User %d entered amount: %s", c.Sender().ID, c.Text())
	return c.Send("<b>Пожалуйста, загрузите фото чека для подтверждения:</b>")
}

func (h *Handler) onUploadReceipt(c tele.Context) error {
	fsm := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	if fsm.Current() != StateEnterAmount {
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Вы не можете загрузить чек сейчас. Начните процесс заново.</b>")
	}

	if c.Message().Photo == nil {
		return c.Send("<b>Пожалуйста, отправьте фото чека.</b>")
	}

	if _, err := createDonation(h, c); err != nil {
		log.Printf("Error saving donation: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Ошибка при сохранении данных пожертвования. Попробуйте позже.</b>")
	}

	if err := fsm.Event(ctx, StateFinish); err != nil {
		log.Printf("FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Произошла ошибка. Начните процесс заново.</b>")
	}

	reaction := tele.Reaction{
		Type:  "emoji",
		Emoji: "🤝",
	}

	reactions := tele.Reactions{
		Reactions: []tele.Reaction{reaction},
		Big:       false,
	}

	if err := c.Bot().React(c.Sender(), c.Message(), reactions); err != nil {
		log.Printf("Ошибка при добавлении реакции: %v", err)
		return c.Send("Не удалось добавить реакцию.")
	}

	h.deleteAllMessages(c)
	h.resetFSM(c.Sender().ID)

	menu := &tele.ReplyMarkup{}
	btnMainBack := menu.Data("⬅️ В главное меню", "main_menu")
	menu.Inline(menu.Row(btnMainBack))

	return c.Send("<b>Спасибо за ваше пожертвование! Ваша поддержка важна.</b>", menu)
}

func (h *Handler) resetFSM(userID int64) {
	log.Printf("Resetting FSM for User %d", userID)
	if h.UserFSM[userID] != nil {
		h.UserFSM[userID].SetState(StateStart)
	}
	delete(h.UserData, userID)
}

// Utility functions
func deleteMessage(c tele.Context) error {
	if c.Callback().Message != nil {
		return c.Bot().Delete(c.Callback().Message)
	}
	return nil
}

func (h *Handler) deleteAllMessages(c tele.Context) {
	if msgs, ok := h.UserData[c.Sender().ID]["messages"].([]*tele.Message); ok {
		for _, msg := range msgs {
			if err := c.Bot().Delete(msg); err != nil {
				log.Printf("Error deleting message: %v", err)
			}
		}
		h.UserData[c.Sender().ID]["messages"] = []*tele.Message{}
	} else {
		log.Printf("No messages found for user %d", c.Sender().ID)
	}
}

func (h *Handler) onMainMenu(c tele.Context) error {
	h.resetFSM(c.Sender().ID)
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	return h.onStart(c)
}

type BankDetails struct {
	BankName string
	Details  string
}

func getBankDetails(bank string) (BankDetails, bool) {
	switch bank {
	case "sber":
		return BankDetails{
			BankName: "Сбербанк",
			Details:  "🟢 <b>Реквизиты Сбербанка:</b>\nКарта:<code> 2202 2080 3701 1005</code>\n<b>Получатель:</b> Лукичёва Ю.Н",
		}, true
	case "vtb":
		return BankDetails{
			BankName: "ВТБ",
			Details:  "🔵 <b>Реквизиты ВТБ:</b>\nКарта:<code> 2200 2402 1368 9108</code>\n<b>Получатель:</b> Лукичёва Ю.Н",
		}, true
	case "sbp":
		return BankDetails{
			BankName: "СБП",
			Details:  "💠 <b>Реквизиты СБП:</b>\nТелефон:<code> +7 963 752-92-99</code>\n<b>Получатель:</b> Лукичёва Ю.Н",
		}, true
	default:
		return BankDetails{}, false
	}
}

func ensureUserData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}
	return data
}

func createBankMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	btnSber := menu.Data("🟢 Сбербанк", "sber", "sber")
	btnVTB := menu.Data("🔵 ВТБ", "vtb", "vtb")
	btnSBP := menu.Data("💠 СБП", "sbp", "sbp")

	menu.Inline(
		menu.Row(btnSber, btnVTB, btnSBP),
	)
	return menu
}

func createDonation(h *Handler, c tele.Context) (models.Donation, error) {
	photo := c.Message().Photo
	fileID := photo.FileID

	var user models.User
	if err := h.DB.FirstOrCreate(&user, models.User{TgID: c.Sender().ID}).Error; err != nil {
		return models.Donation{}, err
	}

	bank, _ := h.UserData[c.Sender().ID]["bank"].(string)
	amountStr, _ := h.UserData[c.Sender().ID]["amount"].(string)
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return models.Donation{}, err
	}

	donation := models.Donation{
		UserID:       user.ID,
		BankName:     bank,
		Amount:       float64(amount),
		ReceiptPhoto: fileID,
	}

	if err := h.DB.Create(&donation).Error; err != nil {
		return models.Donation{}, err
	}

	var total models.TotalDonation
	h.DB.Find(&total)
	total.Total += donation.Amount

	if err := h.DB.Save(&total).Error; err != nil {
		return models.Donation{}, err
	}

	return donation, nil
}
