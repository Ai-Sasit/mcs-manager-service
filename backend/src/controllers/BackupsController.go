package controllers

import (
	"fmt"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
)

type BackupInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// ListBackups lists all zip files in the backup directory for a server
func ListBackups(c fiber.Ctx) error {
	id := c.Params("id")
	if _, ok := state.GetServer(id); !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	backupDir := filepath.Join("data", "backups", id)
	os.MkdirAll(backupDir, 0755)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return utils.ErrorResponse(c, "Failed to read backups: "+err.Error(), fiber.StatusInternalServerError)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".zip" {
			info, _ := entry.Info()
			backups = append(backups, BackupInfo{
				Name:      entry.Name(),
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			})
		}
	}

	return utils.SuccessResponse(c, "OK", backups)
}

// CreateBackup creates a new zip backup of the server directory
func CreateBackup(c fiber.Ctx) error {
	id := c.Params("id")
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	backupDir := filepath.Join("data", "backups", id)
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("%s_%s.zip", srv.Name, timestamp)
	targetPath := filepath.Join(backupDir, backupName)

	go func() {
		utils.LogAudit("system", "BACKUP_START", id, "Starting backup: "+backupName)
		err := utils.CreateZip(srv.ServerDir, targetPath)
		if err != nil {
			utils.LogAudit("system", "BACKUP_FAILED", id, "Backup failed: "+err.Error())
			return
		}
		utils.LogAudit("system", "BACKUP_SUCCESS", id, "Backup completed: "+backupName)
	}()

	return utils.SuccessResponse(c, "Backup task started in background", nil)
}

// DeleteBackup removes a backup file
func DeleteBackup(c fiber.Ctx) error {
	id := c.Params("id")
	name := c.Params("name")

	safeName := filepath.Base(name)
	if safeName == "." || safeName == ".." || safeName == "" {
		return utils.ErrorResponse(c, "Invalid backup name", fiber.StatusBadRequest)
	}

	backupDir := filepath.Join("data", "backups", id)
	filePath := filepath.Join(backupDir, safeName)

	// Security: ensure path stays within backups directory
	cleanPath := filepath.Clean(filePath)
	cleanDir := filepath.Clean(backupDir)
	if len(cleanPath) <= len(cleanDir) || cleanPath[:len(cleanDir)] != cleanDir {
		return utils.ErrorResponse(c, "Invalid backup name", fiber.StatusBadRequest)
	}

	if err := os.Remove(filePath); err != nil {
		return utils.ErrorResponse(c, "Failed to delete backup: "+err.Error(), fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "BACKUP_DELETE", id, "Deleted backup: "+safeName)
	return utils.SuccessResponse(c, "Backup deleted", nil)
}
