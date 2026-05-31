package models

import "github.com/robfig/cron/v3"

type ScheduleEntry struct {
	ID       string       `bson:"id" json:"id"`
	ServerID string       `bson:"server_id" json:"server_id"`
	Task     string       `bson:"task" json:"task"`       // "backup", "restart", "command"
	Cron     string       `bson:"cron" json:"cron"`       // e.g. "0 0 * * *"
	Command  string       `bson:"command" json:"command"` // Only for task="command"
	Enabled  bool         `bson:"enabled" json:"enabled"`
	EntryID  cron.EntryID `bson:"-" json:"-"`
}
