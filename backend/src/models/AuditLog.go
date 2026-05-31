package models

import "time"

type AuditLog struct {
	ID        string    `bson:"id" json:"id"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	User      string    `bson:"user" json:"user"`
	Action    string    `bson:"action" json:"action"`
	Target    string    `bson:"target" json:"target"`
	Details   string    `bson:"details" json:"details"`
}
