package controllers

import (
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"

	"github.com/gofiber/fiber/v3"
)

// ListSchedules returns all configured schedules
func ListSchedules(c fiber.Ctx) error {
	list := state.Scheduler.ListSchedules()
	return utils.SuccessResponse(c, "OK", list)
}

// CreateSchedule adds a new automated task
func CreateSchedule(c fiber.Ctx) error {
	var entry models.ScheduleEntry
	if err := c.Bind().JSON(&entry); err != nil {
		return utils.ErrorResponse(c, "Invalid request body", fiber.StatusBadRequest)
	}

	if entry.ServerID == "" || entry.Task == "" || entry.Cron == "" {
		return utils.ErrorResponse(c, "server_id, task, and cron are required", fiber.StatusBadRequest)
	}

	entry.ID = utils.GenerateID()
	entry.Enabled = true

	if err := state.Scheduler.AddSchedule(&entry); err != nil {
		return utils.ErrorResponse(c, "Failed to schedule task: "+err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit("admin", "SCHEDULE_CREATE", entry.ServerID, "Created automated task: "+entry.Task)
	return utils.SuccessResponse(c, "Task scheduled", entry, fiber.StatusCreated)
}

// DeleteSchedule removes a task
func DeleteSchedule(c fiber.Ctx) error {
	id := c.Params("id")
	state.Scheduler.DeleteSchedule(id)
	utils.LogAudit("admin", "SCHEDULE_DELETE", "system", "Deleted schedule: "+id)
	return utils.SuccessResponse(c, "Deleted", nil)
}

// ToggleSchedule enables or disables a task
func ToggleSchedule(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return utils.ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	if err := state.Scheduler.ToggleSchedule(id, req.Enabled); err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}

	utils.LogAudit("admin", "SCHEDULE_TOGGLE", "system", "Schedule toggled: "+id)
	return utils.SuccessResponse(c, "OK", nil)
}
