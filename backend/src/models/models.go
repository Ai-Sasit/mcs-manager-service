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
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Edition    ServerEdition `json:"edition"`
	ServerType string        `json:"server_type"`
	Version    string        `json:"version"`
	Port       uint16        `json:"port"`
	MaxPlayers uint32        `json:"max_players"`
	MemoryMB   uint32        `json:"memory_mb"`
	Status     ServerStatus  `json:"status"`
	Pid        int           `json:"pid,omitempty"`
	CreatedAt  string        `json:"created_at"`
	ServerDir  string        `json:"server_dir"`
}

type VersionInfo struct {
	ID          string `json:"id"`
	VersionType string `json:"type"`
}

type PluginInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}
