package models

type ServerEdition string
type ServerStatus string

const (
	EditionJava    ServerEdition = "java"
	EditionBedrock ServerEdition = "bedrock"
)

const (
	StatusStopped  ServerStatus = "stopped"
	StatusRunning  ServerStatus = "running"
	StatusStarting ServerStatus = "starting"
)

type ServerConfig struct {
	ID         string        `bson:"id" json:"id"`
	Name       string        `bson:"name" json:"name"`
	Edition    ServerEdition `bson:"edition" json:"edition"`
	ServerType string        `bson:"server_type" json:"server_type"`
	ModLoader  string        `bson:"mod_loader,omitempty" json:"mod_loader,omitempty"`
	Version    string        `bson:"version" json:"version"`
	Port       uint16        `bson:"port" json:"port"`
	MaxPlayers uint32        `bson:"max_players" json:"max_players"`
	MemoryMB   uint32        `bson:"memory_mb" json:"memory_mb"`
	Status     ServerStatus  `bson:"status" json:"status"`
	Pid        int           `bson:"-" json:"pid,omitempty"`
	CreatedAt  string        `bson:"created_at" json:"created_at"`
	ServerDir  string        `bson:"server_dir" json:"server_dir"`
}

type VersionInfo struct {
	ID          string `json:"id"`
	VersionType string `json:"type"`
}

type PluginInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ModInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ProcessInfo struct {
	Pid      int    `json:"pid"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}
