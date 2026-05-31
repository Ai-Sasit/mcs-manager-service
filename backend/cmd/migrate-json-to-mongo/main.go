package main

import (
	"context"
	"encoding/json"
	"fmt"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type appSettings struct {
	AppName          string `bson:"app_name" json:"app_name"`
	DefaultRAM       int    `bson:"default_ram" json:"default_ram"`
	DiscordWebhook   string `bson:"discord_webhook" json:"discord_webhook"`
	TelemetryEnabled bool   `bson:"telemetry_enabled" json:"telemetry_enabled"`
}

type settingsDocument struct {
	Key         string `bson:"key"`
	appSettings `bson:",inline"`
}

func main() {
	_ = godotenv.Load()
	db := utils.ConnectDB()
	defer utils.DisconnectDB(db)

	dataDir := filepath.Join(".", "data")
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}

	fmt.Printf("[migrate] Importing JSON data from %s\n", dataDir)
	importServers(db, filepath.Join(dataDir, "servers.json"))
	importUsers(db, filepath.Join(dataDir, "users.json"))
	importSchedules(db, filepath.Join(dataDir, "schedules.json"))
	importSettings(db, filepath.Join(dataDir, "settings.json"))
	importAuditLogs(db, filepath.Join(dataDir, "audit.json"))
	fmt.Println("[migrate] Done")
}

func readJSON(path string, out any) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("[migrate] Skip %s: %v\n", path, err)
		return false
	}
	if err := json.Unmarshal(data, out); err != nil {
		fmt.Printf("[migrate] Invalid JSON %s: %v\n", path, err)
		return false
	}
	return true
}

func upsertByID(col *mongo.Collection, id string, doc any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := col.ReplaceOne(ctx, bson.M{"id": id}, doc, options.Replace().SetUpsert(true))
	return err
}

func importServers(db *mongo.Client, path string) {
	servers := map[string]*models.ServerConfig{}
	if !readJSON(path, &servers) {
		return
	}
	col := utils.GetCollection(db, "servers")
	count := 0
	for key, server := range servers {
		if server.ID == "" {
			server.ID = key
		}
		server.Status = models.StatusStopped
		if err := upsertByID(col, server.ID, server); err != nil {
			fmt.Printf("[migrate] Server %s failed: %v\n", server.ID, err)
			continue
		}
		count++
	}
	fmt.Printf("[migrate] Imported %d server(s)\n", count)
}

func importUsers(db *mongo.Client, path string) {
	users := map[string]*models.User{}
	if !readJSON(path, &users) {
		return
	}
	col := utils.GetCollection(db, "users")
	count := 0
	for key, user := range users {
		if user.ID == "" {
			user.ID = key
		}
		if err := upsertByID(col, user.ID, user); err != nil {
			fmt.Printf("[migrate] User %s failed: %v\n", user.ID, err)
			continue
		}
		count++
	}
	fmt.Printf("[migrate] Imported %d user(s)\n", count)
}

func importSchedules(db *mongo.Client, path string) {
	var schedules []models.ScheduleEntry
	if !readJSON(path, &schedules) {
		return
	}
	col := utils.GetCollection(db, "schedules")
	count := 0
	for _, schedule := range schedules {
		if schedule.ID == "" {
			schedule.ID = utils.GenerateID()
		}
		if err := upsertByID(col, schedule.ID, schedule); err != nil {
			fmt.Printf("[migrate] Schedule %s failed: %v\n", schedule.ID, err)
			continue
		}
		count++
	}
	fmt.Printf("[migrate] Imported %d schedule(s)\n", count)
}

func importSettings(db *mongo.Client, path string) {
	var settings appSettings
	if !readJSON(path, &settings) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc := settingsDocument{Key: "global", appSettings: settings}
	_, err := utils.GetCollection(db, "settings").ReplaceOne(ctx, bson.M{"key": "global"}, doc, options.Replace().SetUpsert(true))
	if err != nil {
		fmt.Printf("[migrate] Settings failed: %v\n", err)
		return
	}
	fmt.Println("[migrate] Imported settings")
}

func importAuditLogs(db *mongo.Client, path string) {
	var logs []models.AuditLog
	if !readJSON(path, &logs) {
		return
	}
	col := utils.GetCollection(db, "audit_logs")
	count := 0
	for _, log := range logs {
		if log.ID == "" {
			log.ID = utils.GenerateID()
		}
		if err := upsertByID(col, log.ID, log); err != nil {
			fmt.Printf("[migrate] Audit log %s failed: %v\n", log.ID, err)
			continue
		}
		count++
	}
	fmt.Printf("[migrate] Imported %d audit log(s)\n", count)
}
