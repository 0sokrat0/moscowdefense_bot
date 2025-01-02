package handlers

import (
	"log"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onText(c tele.Context) error {
	if c.Chat().Type != tele.ChatPrivate {
		return nil // Игнорируем сообщения из групп или каналов
	}

	mode, _ := h.UserData[c.Sender().ID]["mode"].(string)

	// Если админ
	if mode == "admin" {
		adminFSM, exists := h.AdminFSM[c.Sender().ID]
		if !exists {
			h.resetFSM(c.Sender().ID)
			return c.Send("Нет активного админского процесса. Попробуйте снова.")
		}

		switch adminFSM.Current() {
		// --- Добавление цели ---
		case StateAddGoalTitle:
			return h.processGoalTitle(c, adminFSM)
		case StateAddGoalDescription:
			return h.processGoalDescription(c, adminFSM)
		case StateAddGoalTargetSum:
			return h.processGoalTargetSum(c, adminFSM)

		// --- Баланс ---
		case "add_funds":
			return h.processAddFunds(c, adminFSM)
		case "sub_funds":
			return h.processSubFunds(c, adminFSM)
		case "adding_funds":
			return h.processAddFunds(c, adminFSM)
		case "subtracting_funds":
			return h.processSubFunds(c, adminFSM)

		// --- Редактирование цели ---
		case StateEditGoalWaitInput:
			return h.onTextAdminEdit(c, adminFSM)

		// --- Добавление администратора (то, что не хватает!) ---
		case StateAddAdminWaitID:
			// Здесь обрабатываем ввод TG ID
			return h.processNewAdminID(c, adminFSM)

		case StateAddAdminWaitUsername:
			// Здесь обрабатываем ввод username
			return h.processNewAdminUsername(c, adminFSM)

		default:
			// Неизвестное состояние
			log.Printf("[WARN] Неизвестное состояние FSM администратора ID=%d: %s",
				c.Sender().ID, adminFSM.Current())
			h.resetFSM(c.Sender().ID)
			return c.Send("Неверный ввод. Начните заново или используйте /panel.")
		}

	} else if mode == "user" {
		// ... если режим user ...
		userFSM, exists := h.UserFSM[c.Sender().ID]
		if !exists {
			h.resetFSM(c.Sender().ID)
			return c.Send("Нет активного процесса. Введите /start, чтобы начать.")
		}

		switch userFSM.Current() {
		case StateSelectBank:
			return h.onBankDetails(c)
		case StateEnterAmount:
			return h.onEnterAmount(c)
		default:
			log.Printf("[WARN] Неизвестное состояние FSM пользователя ID=%d: %s",
				c.Sender().ID, userFSM.Current())
			h.resetFSM(c.Sender().ID)
			return c.Send("Неверный ввод. Начните заново.")
		}
	}

	// Если режим не установлен
	h.resetFSM(c.Sender().ID)
	return c.Send("У вас нет активного процесса. Введите /start, чтобы начать.")
}
