package controllers

import (
	"mc-manage-backend/src/utils"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AppSettings struct {
	AppName          string `bson:"app_name" json:"app_name"`
	DefaultRAM       int    `bson:"default_ram" json:"default_ram"`
	DiscordWebhook   string `bson:"discord_webhook" json:"discord_webhook"`
	TelemetryEnabled bool   `bson:"telemetry_enabled" json:"telemetry_enabled"`
}

type settingsDocument struct {
	Key         string `bson:"key"`
	AppSettings `bson:",inline"`
}

func defaultSettings() AppSettings {
	return AppSettings{
		AppName:    "MC Management Dashboard",
		DefaultRAM: 2048,
	}
}

func settingsCollection() *mongo.Collection {
	return utils.GetCollection(utils.DB, "settings")
}

func loadSettings() AppSettings {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()

	var doc settingsDocument
	err := settingsCollection().FindOne(ctx, bson.M{"key": "global"}).Decode(&doc)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("[Settings] Failed to load settings: "+err.Error(), nil)
		}
		return defaultSettings()
	}
	return doc.AppSettings
}

func GetSettings(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    loadSettings(),
	})
}

func UpdateSettings(c fiber.Ctx) error {
	var newSettings AppSettings
	if err := c.Bind().JSON(&newSettings); err != nil {
		return utils.ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()

	doc := settingsDocument{Key: "global", AppSettings: newSettings}
	if _, err := settingsCollection().ReplaceOne(ctx, bson.M{"key": "global"}, doc, options.Replace().SetUpsert(true)); err != nil {
		logger.Error("[Settings] Failed to save settings: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Failed to save settings", fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "SETTINGS_UPDATE", "system", "Updated global application settings.")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Settings updated",
	})
}
