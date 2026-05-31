package models

type AppSettings struct {
	AppName          string `bson:"app_name" json:"app_name"`
	DefaultRAM       int    `bson:"default_ram" json:"default_ram"`
	DiscordWebhook   string `bson:"discord_webhook" json:"discord_webhook"`
	TelemetryEnabled bool   `bson:"telemetry_enabled" json:"telemetry_enabled"`
}

type SettingsDocument struct {
	Key         string `bson:"key"`
	AppSettings `bson:",inline"`
}

func DefaultSettings() AppSettings {
	return AppSettings{
		AppName:    "MC Management Dashboard",
		DefaultRAM: 2048,
	}
}
