package controllers

import (
	"context"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func settingsCollection() *mongo.Collection {
	return utils.GetCollection(utils.DB, "settings")
}

func loadSettings() models.AppSettings {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc models.SettingsDocument
	err := settingsCollection().FindOne(ctx, bson.M{"key": "global"}).Decode(&doc)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("[Settings] Failed to load settings: "+err.Error(), nil)
		}
		return models.DefaultSettings()
	}
	return doc.AppSettings
}

func GetSettings(c fiber.Ctx) error {
	return utils.SuccessResponse(c, "OK", loadSettings())
}

func UpdateSettings(c fiber.Ctx) error {
	var newSettings models.AppSettings
	if err := c.Bind().JSON(&newSettings); err != nil {
		return utils.ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	doc := models.SettingsDocument{Key: "global", AppSettings: newSettings}
	if _, err := settingsCollection().ReplaceOne(c.Context(), bson.M{"key": "global"}, doc, options.Replace().SetUpsert(true)); err != nil {
		logger.Error("[Settings] Failed to save settings: "+err.Error(), nil)
		return utils.ErrorResponse(c, "Failed to save settings", fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "SETTINGS_UPDATE", "system", "Updated global application settings.")

	return utils.SuccessResponse(c, "Settings updated", nil)
}
