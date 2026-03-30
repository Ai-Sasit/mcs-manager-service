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

	return c.JSON(fiber.Map{
		"success": true,
		"data":    backups,
	})
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

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Backup task started in background",
	})
}

// DeleteBackup removes a backup file
func DeleteBackup(c fiber.Ctx) error {
	id := c.Params("id")
	name := c.Params("name")

	path := filepath.Join("data", "backups", id, name)
	if err := os.Remove(path); err != nil {
		return utils.ErrorResponse(c, "Failed to delete backup: "+err.Error(), fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "BACKUP_DELETE", id, "Deleted backup: "+name)
	return c.JSON(fiber.Map{"success": true, "message": "Backup deleted"})
}
