package interfaces

import "mc-manage-backend/src/models"

type CreateServerRequest struct {
	Name       string               `json:"name"`
	Edition    models.ServerEdition `json:"edition"`
	Version    string               `json:"version"`
	Port       *uint16              `json:"port,omitempty"`
	MaxPlayers *uint32              `json:"max_players,omitempty"`
	MemoryMB   *uint32              `json:"memory_mb,omitempty"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}
