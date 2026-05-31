package controllers

import (
	"io"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
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

// UploadPlugin handles plugin file upload
func UploadPlugin(c fiber.Ctx) error {
	id := c.Params("id")
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

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, "No file uploaded", fiber.StatusBadRequest)
	}

	filename := filepath.Base(file.Filename)
	destPath := filepath.Join(pluginsDir, filename)

	src, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, "Failed to open uploaded file", fiber.StatusInternalServerError)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return utils.ErrorResponse(c, "Failed to create destination file", fiber.StatusInternalServerError)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return utils.ErrorResponse(c, "Failed to save file", fiber.StatusInternalServerError)
	}

	utils.LogAudit("admin", "PLUGIN_UPLOAD", id, "Uploaded plugin: "+filename)
	return utils.SuccessResponse(c, "Plugin uploaded", nil)
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
