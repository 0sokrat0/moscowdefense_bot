package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"TgDonation/internal/database/models"
)

func (h *Handler) sendThankYouToGroup(user tele.User, amount float64) {
	username := user.Username
	firstName := user.FirstName

	// Формируем имя пользователя
	var displayName string
	if username != "" {
		displayName = "@" + username
	} else {
		displayName = firstName
	}

	// Формируем текст сообщения
	message := fmt.Sprintf(
		"🙏 Спасибо, <b>%s</b>, за помощь фронту в размере <b>%s ₽</b>!",
		displayName,
		formatFloatNoTrailingZeros(amount),
	)

	// Отправляем сообщение в группу с использованием HTML-разметки
	_, err := h.Bot.Send(tele.ChatID(h.GroupChatID), message, &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
	if err != nil {
		log.Printf("Ошибка отправки благодарственного сообщения в группу: %v", err)
	} else {
		log.Printf("Благодарственное сообщение отправлено в группу: %s", message)
	}
}

// onDonation инициализирует процесс пожертвования
func (h *Handler) onDonation(c tele.Context) error {
	// 1. Сбрасываем FSM (и если нужно, режим)
	h.resetFSM(c.Sender().ID)
	fsmObj := h.getOrCreateFSM(c.Sender().ID)

	// 2. Убеждаемся, что сейчас мы в состоянии StateStart
	if fsmObj.Current() != StateStart {
		return c.Send("Вы не можете начать процесс пожертвования в текущем состоянии.")
	}

	// 3. Удаляем сообщение обратного вызова (если есть)
	if err := deleteMessage(c); err != nil {
		log.Printf("[onDonation] Ошибка удаления сообщения: %v", err)
	}

	// 4. Устанавливаем режим user
	h.UserData[c.Sender().ID] = ensureUserData(h.UserData[c.Sender().ID])
	h.UserData[c.Sender().ID]["mode"] = "user"

	// 5. Переходим в состояние выбора банка (StateSelectBank)
	ctx := context.Background()
	if err := fsmObj.Event(ctx, StateSelectBank); err != nil {
		log.Printf("[onDonation] FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка FSM. Начните процесс заново, выбрав \"Сделать пожертвование\".")
	}

	// 6. Показываем меню с банками
	menu := createBankMenu()
	return c.Send("<b>Выберите банк для пожертвования:</b>", menu)
}

// onBankDetails обрабатывает выбор банка (sber, vtb, sbp)
func (h *Handler) onBankDetails(c tele.Context) error {
	fsmObj := h.getOrCreateFSM(c.Sender().ID)

	// Удаляем старое сообщение (если оно callback)
	if err := deleteMessage(c); err != nil {
		log.Printf("[onBankDetails] Ошибка удаления сообщения: %v", err)
	}

	// Проверяем, что мы действительно в состоянии StateSelectBank
	if fsmObj.Current() != StateSelectBank {
		h.resetFSM(c.Sender().ID)
		return c.Send("Выбор банка сейчас недоступен. Начните процесс заново.")
	}

	// Сохраняем выбранный банк
	selectedBank := c.Callback().Data
	bankDetails, valid := getBankDetails(selectedBank)
	if !valid {
		return c.Send("Неизвестный банк. Попробуйте снова.")
	}

	h.UserData[c.Sender().ID] = ensureUserData(h.UserData[c.Sender().ID])
	h.UserData[c.Sender().ID]["bank"] = bankDetails.BankName

	// Переходим в StateEnterAmount
	ctx := context.Background()
	if err := fsmObj.Event(ctx, StateEnterAmount); err != nil {
		log.Printf("[onBankDetails] FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("Произошла ошибка FSM. Начните процесс заново.")
	}

	// Выводим реквизиты банка и просим ввести сумму
	text := bankDetails.Details + "\n\n<b>Введите сумму пожертвования:</b>"
	return c.Send(text)
}

// onEnterAmount обрабатывает ввод пользователем суммы
func (h *Handler) onEnterAmount(c tele.Context) error {
	fsmObj := h.getOrCreateFSM(c.Sender().ID)

	// Проверяем текущее состояние
	if fsmObj.Current() != StateEnterAmount {
		h.resetFSM(c.Sender().ID)
		return c.Send("Вы не можете ввести сумму сейчас. Начните процесс заново.")
	}

	// Попробуем добавить реакцию «👌»
	reaction := tele.Reaction{Type: "emoji", Emoji: "👌"}
	reactions := tele.Reactions{Reactions: []tele.Reaction{reaction}, Big: false}
	if err := c.Bot().React(c.Sender(), c.Message(), reactions); err != nil {
		log.Printf("[onEnterAmount] Ошибка при добавлении реакции: %v", err)
		// не выходим, просто логируем
	}

	// Сохраняем сумму (как строку)
	userData := ensureUserData(h.UserData[c.Sender().ID])
	userData["amount"] = c.Text()

	// Проверим, что это число
	amountValue, err := strconv.ParseFloat(c.Text(), 64)
	if err != nil || amountValue <= 0 {
		return c.Send("<b>Введите корректную сумму (числом).</b>")
	}

	log.Printf("[onEnterAmount] User %d entered amount: %s", c.Sender().ID, c.Text())

	// Просим загрузить чек
	return c.Send("<b>Пожалуйста, загрузите фото(скрин) чека для подтверждения:</b>")
}

// onUploadReceipt обрабатывает загрузку фотографии
func (h *Handler) onUploadReceipt(c tele.Context) error {
	fsmObj := h.getOrCreateFSM(c.Sender().ID)
	ctx := context.Background()

	// Проверяем состояние
	if fsmObj.Current() != StateEnterAmount {
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Вы не можете загрузить чек сейчас. Начните процесс заново.</b>")
	}

	// Проверяем, что пользователь отправил фото
	if c.Message().Photo == nil {
		return c.Send("<b>Пожалуйста, отправьте фотографию чека.</b>")
	}

	// Создаём пожертвование в БД
	donation, err := createDonation(h, c)
	if err != nil {
		log.Printf("[onUploadReceipt] Ошибка при сохранении donation: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Ошибка при сохранении данных пожертвования. Попробуйте позже.</b>")
	}
	log.Printf("[onUploadReceipt] Donation created: ID=%d Amount=%.2f", donation.ID, donation.Amount)

	// Переходим в StateFinish
	if err := fsmObj.Event(ctx, StateFinish); err != nil {
		log.Printf("[onUploadReceipt] FSM Event Error: %v", err)
		h.resetFSM(c.Sender().ID)
		return c.Send("<b>Произошла ошибка FSM. Начните процесс заново.</b>")
	}

	// Отправляем благодарственное сообщение в группу
	h.sendThankYouToGroup(*c.Sender(), donation.Amount)

	// Сбрасываем FSM, т.к. закончили процесс
	h.resetFSM(c.Sender().ID)

	// Предлагаем вернуться в меню
	menu := &tele.ReplyMarkup{}
	btnMainBack := menu.Data("⬅️ В главное меню", "main_menu")
	menu.Inline(menu.Row(btnMainBack))

	return c.Send("<b>Благодарим вас за помощь фронту!</b>", menu)
}

// createDonation – создаёт запись о пожертвовании и обновляет общий баланс.
func createDonation(h *Handler, c tele.Context) (models.Donation, error) {
	// 1. Получаем пользователя
	var user models.User
	if err := h.DB.FirstOrCreate(&user, models.User{TgID: c.Sender().ID}).Error; err != nil {
		return models.Donation{}, fmt.Errorf("ошибка при создании/поиске пользователя: %w", err)
	}

	// 2. Получаем поля bank и amount
	data := ensureUserData(h.UserData[c.Sender().ID])
	bank, ok := data["bank"].(string)
	if !ok || bank == "" {
		return models.Donation{}, fmt.Errorf("не указан банк в userData")
	}

	amountStr, ok := data["amount"].(string)
	if !ok || amountStr == "" {
		return models.Donation{}, fmt.Errorf("не указана сумма в userData")
	}

	amountValue, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amountValue <= 0 {
		return models.Donation{}, fmt.Errorf("сумма невалидна: %w", err)
	}

	// 3. Получаем FileID из фотки чека
	photo := c.Message().Photo
	if photo == nil {
		return models.Donation{}, fmt.Errorf("нет фотографии в сообщении")
	}
	fileID := photo.FileID

	// 4. Создаём Donation
	donation := models.Donation{
		UserID:       user.ID,
		BankName:     bank,
		Amount:       amountValue,
		ReceiptPhoto: fileID,
	}
	// Пытаемся сохранить
	if err := h.DB.Create(&donation).Error; err != nil {
		return models.Donation{}, fmt.Errorf("ошибка сохранения donation: %w", err)
	}

	log.Printf("[createDonation] Donation (ID=%d) saved. Updating total balance...", donation.ID)

	// 5. Обновляем общий баланс
	var total models.TotalDonation
	if err := h.DB.First(&total).Error; err != nil {
		// Если нет записи, создадим
		if err.Error() == "record not found" {
			total.Total = donation.Amount
			if err2 := h.DB.Create(&total).Error; err2 != nil {
				return models.Donation{}, fmt.Errorf("ошибка при создании totalDonation: %w", err2)
			}
		} else {
			return models.Donation{}, fmt.Errorf("ошибка поиска totalDonation: %w", err)
		}
	} else {
		// Запись есть, обновим
		total.Total += donation.Amount
		if err := h.DB.Save(&total).Error; err != nil {
			return models.Donation{}, fmt.Errorf("ошибка обновления totalDonation: %w", err)
		}
	}
	log.Printf("[createDonation] Общий баланс обновлён: %.2f", total.Total)

	return donation, nil
}

// deleteMessage удаляет callback-сообщение, если есть
func deleteMessage(c tele.Context) error {
	if c.Callback() != nil && c.Callback().Message != nil {
		return c.Bot().Delete(c.Callback().Message)
	}
	return nil
}

// ensureUserData убеждается, что data не nil
func ensureUserData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}
	return data
}

// getBankDetails возвращает реквизиты банка по его короткому имени (sber, vtb, sbp)
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

// createBankMenu создаёт клавиатуру с выбором банка
func createBankMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	btnSber := menu.Data("🟢 Сбербанк", "sber", "sber")
	btnVTB := menu.Data("🔵 ВТБ", "vtb", "vtb")
	btnSBP := menu.Data("💠 СБП", "sbp", "sbp")
	btnBack := menu.Data("⬅️ Назад", "main_menu")

	menu.Inline(
		menu.Row(btnSber, btnVTB, btnSBP),
		menu.Row(btnBack),
	)
	return menu
}

// BankDetails хранит краткое имя банка и текст с реквизитами
type BankDetails struct {
	BankName string
	Details  string
}

func (h *Handler) onMainMenu(c tele.Context) error {
	h.resetFSM(c.Sender().ID)
	if err := c.Bot().Delete(c.Callback().Message); err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}

	return h.onStart(c)
}
