package handlers

import (
	"context"
	"strconv"

	"TgDonation/internal/database/models"
	"github.com/looplab/fsm"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) onAddAdminStart(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("У вас нет доступа к этой функции.")
	}

	h.resetFSM(c.Sender().ID)
	fsmObj := h.getOrCreateAdminFSM(c.Sender().ID)
	h.UserData[c.Sender().ID] = map[string]interface{}{
		"action": "add_admin",
		"mode":   "admin",
	}

	if err := fsmObj.Event(context.Background(), "add_admin_id"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Попробуйте снова.")
	}

	return c.Send("Введите TG ID нового администратора (числом):")
}

func (h *Handler) processNewAdminID(c tele.Context, fsmObj *fsm.FSM) error {
	idStr := c.Text()
	newAdminID, err := strconv.Atoi(idStr)
	if err != nil || newAdminID <= 0 {
		return c.Send("Некорректный ID. Введите положительное число.")
	}

	h.UserData[c.Sender().ID]["new_admin_id"] = newAdminID

	if err := fsmObj.Event(context.Background(), "add_admin_username"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка. Попробуйте снова.")
	}

	return c.Send("Введите username нового администратора (без @):")
}

func (h *Handler) processNewAdminUsername(c tele.Context, fsmObj *fsm.FSM) error {
	username := c.Text()
	if username == "" {
		return c.Send("Username не может быть пустым.")
	}

	h.UserData[c.Sender().ID]["new_admin_username"] = username

	if err := fsmObj.Event(context.Background(), "finish_add_admin"); err != nil {
		h.resetFSM(c.Sender().ID)
		return c.Send("Ошибка при добавлении администратора. Попробуйте снова.")
	}

	return h.finishAddAdmin(c)
}

func (h *Handler) finishAddAdmin(c tele.Context) error {
	data := h.UserData[c.Sender().ID]
	if data == nil {
		return c.Send("Нет данных для добавления администратора.")
	}

	newAdminID := data["new_admin_id"].(int)
	newAdminUsername := data["new_admin_username"].(string)

	var count int64
	h.DB.Model(&models.Admin{}).Where("tg_id = ?", newAdminID).Count(&count)
	if count > 0 {
		h.resetFSM(c.Sender().ID)
		delete(h.UserData, c.Sender().ID)
		return c.Send("Администратор с таким TG ID уже существует.", backButton("back_to_panel"))
	}

	admin := models.Admin{
		TgID:     int64(newAdminID),
		Username: newAdminUsername,
		Role:     "admin",
	}

	if err := h.DB.Create(&admin).Error; err != nil {
		h.resetFSM(c.Sender().ID)
		delete(h.UserData, c.Sender().ID)
		return c.Send("Ошибка при добавлении администратора.", backButton("back_to_panel"))
	}

	h.resetFSM(c.Sender().ID)
	delete(h.UserData, c.Sender().ID)
	return c.Send("✅ Администратор успешно добавлен!", backButton("back_to_panel"))
}
