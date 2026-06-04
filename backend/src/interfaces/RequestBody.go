package interfaces

import "mc-manage-backend/src/models"

type CreateServerRequest struct {
	Name       string               `json:"name" validate:"required"`
	Edition    models.ServerEdition `json:"edition" validate:"required"`
	ServerType string               `json:"server_type"`
	ModLoader  string               `json:"mod_loader"`
	Version    string               `json:"version" validate:"required"`
	Port       *uint16              `json:"port,omitempty"`
	MaxPlayers *uint32              `json:"max_players,omitempty"`
	MemoryMB   *uint32              `json:"memory_mb,omitempty"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin viewer"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=32"`
	Password string `json:"password" validate:"omitempty,min=8"`
	Role     string `json:"role" validate:"omitempty,oneof=admin viewer"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
