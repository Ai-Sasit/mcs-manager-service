package controllers

import (
	"fmt"
	"io"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

// ListPlugins returns the list of plugins/addons for a server
func ListPlugins(c fiber.Ctx) error {
	id := c.Params("id")
	plugins, err := services.ListPlugins(state, id)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}
	return utils.SuccessResponse(c, "OK", plugins)
}

// UploadPlugin handles plugin file upload (single or multiple)
func UploadPlugin(c fiber.Ctx) error {
	id := c.Params("id")

	// Check server exists
	srv, ok := state.GetServer(id)
	if !ok {
		return utils.ErrorResponse(c, "Server not found", fiber.StatusNotFound)
	}

	subDir := "plugins"
	if srv.Edition == models.EditionBedrock {
		subDir = "addons"
	}
	pluginsDir := filepath.Join(srv.ServerDir, subDir)
	os.MkdirAll(pluginsDir, os.ModePerm)

	// Try multipart form for multiple files
	form, err := c.MultipartForm()
	if err == nil && form != nil && form.File != nil {
		var uploaded []map[string]string
		var errors []map[string]string

		for _, headers := range form.File {
			for _, fh := range headers {
				name, saveErr := saveFileHeader(fh, pluginsDir, id)
				if saveErr != nil {
					errors = append(errors, map[string]string{
						"name":  fh.Filename,
						"error": saveErr.Error(),
					})
				} else {
					uploaded = append(uploaded, map[string]string{"name": name})
				}
			}
		}

		utils.LogAudit("admin", "PLUGIN_UPLOAD", id, fmt.Sprintf("Uploaded %d plugin(s)", len(uploaded)))
		return utils.SuccessResponse(c, "Plugins processed", map[string]interface{}{
			"uploaded": uploaded,
			"errors":   errors,
		})
	}

	// Fallback: single file (backward compat)
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, "No file uploaded", fiber.StatusBadRequest)
	}

	filename, err := saveFileHeader(file, pluginsDir, id)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "PLUGIN_UPLOAD", id, "Uploaded plugin: "+filename)
	return utils.SuccessResponse(c, "Plugin uploaded", nil)
}

// saveFileHeader saves a single multipart file header to the plugins directory.
func saveFileHeader(fh *multipart.FileHeader, pluginsDir, serverID string) (string, error) {
	filename := filepath.Base(fh.Filename)
	if filename == "." || filename == ".." {
		return "", fmt.Errorf("invalid filename")
	}

	destPath := filepath.Join(pluginsDir, filename)

	// Security: ensure path stays within plugins dir
	cleanPath := filepath.Clean(destPath)
	cleanDir := filepath.Clean(pluginsDir)
	if len(cleanPath) <= len(cleanDir) || cleanPath[:len(cleanDir)] != cleanDir {
		return "", fmt.Errorf("invalid filename")
	}

	src, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file")
	}

	utils.LogAudit("admin", "PLUGIN_UPLOAD", serverID, "Uploaded plugin: "+filename)
	return filename, nil
}

// DeletePlugin removes a plugin file
func DeletePlugin(c fiber.Ctx) error {
	id := c.Params("id")
	name := c.Params("name")
	if err := services.DeletePlugin(state, id, name); err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}
	utils.LogAudit("admin", "PLUGIN_DELETE", id, "Deleted plugin: "+name)
	return utils.SuccessResponse(c, "Deleted", nil)
}
