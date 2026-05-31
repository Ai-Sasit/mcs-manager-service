package utils

import (
	"crypto/rand"
	"encoding/hex"
	"mc-manage-backend/src/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func LogAudit(user, action, target, details string) {
	entry := models.AuditLog{
		ID:        GenerateID(),
		Timestamp: time.Now(),
		User:      user,
		Action:    action,
		Target:    target,
		Details:   details,
	}

	ctx, cancel := MongoContext(10 * time.Second)
	defer cancel()
	if _, err := GetCollection(DB, "audit_logs").InsertOne(ctx, entry); err != nil {
		logger.Error("[Audit] Failed to write audit log: "+err.Error(), nil)
	}
}

func GetAuditLogs() []models.AuditLog {
	ctx, cancel := MongoContext(10 * time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(1000)
	cursor, err := GetCollection(DB, "audit_logs").Find(ctx, bson.M{}, opts)
	if err != nil {
		logger.Error("[Audit] Failed to load audit logs: "+err.Error(), nil)
		return []models.AuditLog{}
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		logger.Error("[Audit] Failed to decode audit logs: "+err.Error(), nil)
		return []models.AuditLog{}
	}
	if logs == nil {
		return []models.AuditLog{}
	}
	return logs
}
