package handlers

import (
	"github.com/looplab/fsm"
)

func (h *Handler) getOrCreateFSM(userID int64) *fsm.FSM {
	if _, exists := h.UserFSM[userID]; !exists {
		h.UserFSM[userID] = fsm.NewFSM(
			StateStart,
			fsm.Events{
				{Name: "bank", Src: []string{StateStart}, Dst: StateSelectBank},
				{Name: "amount", Src: []string{StateSelectBank}, Dst: StateEnterAmount},
				{Name: "finish", Src: []string{StateEnterAmount}, Dst: StateFinish},
			},
			fsm.Callbacks{},
		)
	}
	return h.UserFSM[userID]
}

func (h *Handler) getOrCreateAdminFSM(userID int64) *fsm.FSM {
	if _, exists := h.AdminFSM[userID]; !exists {
		h.AdminFSM[userID] = fsm.NewFSM(
			StateStart,
			fsm.Events{
				// Добавление цели
				{Name: "add_goal_title", Src: []string{StateStart}, Dst: StateAddGoalTitle},
				{Name: "add_goal_description", Src: []string{StateAddGoalTitle}, Dst: StateAddGoalDescription},
				{Name: "add_goal_target_sum", Src: []string{StateAddGoalDescription}, Dst: StateAddGoalTargetSum},
				{Name: "finish_goal", Src: []string{StateAddGoalTargetSum}, Dst: StateFinishedGoal},

				// Добавление администратора
				{Name: "add_admin_id", Src: []string{StateStart}, Dst: StateAddAdminWaitID},
				{Name: "add_admin_username", Src: []string{StateAddAdminWaitID}, Dst: StateAddAdminWaitUsername},
				{Name: "finish_add_admin", Src: []string{StateAddAdminWaitUsername}, Dst: StateFinish},

				// Редактирование цели
				{Name: "go_edit_goal_select", Src: []string{StateStart}, Dst: StateEditGoalSelect},
				{Name: "go_edit_goal_field", Src: []string{StateEditGoalSelect}, Dst: StateEditGoalFieldSelect},
				{Name: "wait_input", Src: []string{StateEditGoalFieldSelect}, Dst: StateEditGoalWaitInput},

				// *** Добавляем события для баланса ***
				{Name: "wait_add_funds_amount", Src: []string{StateStart}, Dst: "adding_funds"},
				{Name: "wait_sub_funds_amount", Src: []string{StateStart}, Dst: "subtracting_funds"},
			},
			fsm.Callbacks{},
		)

	}
	return h.AdminFSM[userID]
}
