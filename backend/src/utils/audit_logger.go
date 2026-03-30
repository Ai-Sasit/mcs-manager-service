package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditLog struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Details   string    `json:"details"`
}

var (
	auditLogs []AuditLog
	auditMu   sync.Mutex
	auditPath = filepath.Join("data", "audit.json")
)

func init() {
	os.MkdirAll("data", 0755)
	file, err := os.ReadFile(auditPath)
	if err == nil {
		json.Unmarshal(file, &auditLogs)
	}
}

func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func LogAudit(user, action, target, details string) {
	auditMu.Lock()
	defer auditMu.Unlock()
	entry := AuditLog{
		ID:        GenerateID(),
		Timestamp: time.Now(),
		User:      user,
		Action:    action,
		Target:    target,
		Details:   details,
	}
	// Prepend for newest first
	auditLogs = append([]AuditLog{entry}, auditLogs...)
	// Keep only last 1000 items
	if len(auditLogs) > 1000 {
		auditLogs = auditLogs[:1000]
	}
	saveAuditLogs()
}

func GetAuditLogs() []AuditLog {
	auditMu.Lock()
	defer auditMu.Unlock()
	if auditLogs == nil {
		return []AuditLog{}
	}
	return auditLogs
}

func saveAuditLogs() {
	data, _ := json.MarshalIndent(auditLogs, "", "  ")
	os.WriteFile(auditPath, data, 0644)
}
