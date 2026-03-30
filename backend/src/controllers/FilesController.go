package controllers

import (
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type FileNode struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

func getSafePath(c fiber.Ctx, id, subPath string) (string, error) {
	srv, ok := state.GetServer(id)
	if !ok {
		return "", fiber.NewError(fiber.StatusNotFound, "Server not found")
	}

	baseDir := filepath.Clean(srv.ServerDir)
	targetPath := filepath.Clean(filepath.Join(baseDir, subPath))

	// Ensure the target is inside the baseDir
	if !strings.HasPrefix(targetPath, baseDir) {
		return "", fiber.NewError(fiber.StatusForbidden, "Invalid path")
	}
	return targetPath, nil
}

// ListFiles lists files in a directory
func ListFiles(c fiber.Ctx) error {
	id := c.Params("id")
	dirPath := c.Query("path", "")

	targetPath, err := getSafePath(c, id, dirPath)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden)
	}

	entries, readErr := os.ReadDir(targetPath)
	if readErr != nil {
		return utils.ErrorResponse(c, readErr.Error(), fiber.StatusInternalServerError)
	}

	var nodes []FileNode
	for _, entry := range entries {
		info, _ := entry.Info()
		nodes = append(nodes, FileNode{
			Name:  entry.Name(),
			Path:  filepath.ToSlash(filepath.Join(dirPath, entry.Name())),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nodes,
	})
}

// ReadFile outputs the content of a file
func ReadFile(c fiber.Ctx) error {
	id := c.Params("id")
	filePath := c.Query("path")

	targetPath, err := getSafePath(c, id, filePath)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden)
	}

	data, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		return utils.ErrorResponse(c, readErr.Error(), fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    string(data),
	})
}

// WriteFileReq represents a file write update
type WriteFileReq struct {
	Content string `json:"content"`
}

// WriteFile writes the given content to a file
func WriteFile(c fiber.Ctx) error {
	id := c.Params("id")
	filePath := c.Query("path")

	targetPath, err := getSafePath(c, id, filePath)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden)
	}

	var req WriteFileReq
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return utils.ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	if writeErr := os.WriteFile(targetPath, []byte(req.Content), 0644); writeErr != nil {
		return utils.ErrorResponse(c, writeErr.Error(), fiber.StatusInternalServerError)
	}
	utils.LogAudit("admin", "FILE_UPDATED", id, "Updated file: "+filePath)

	return c.JSON(fiber.Map{"success": true, "message": "File saved"})
}
