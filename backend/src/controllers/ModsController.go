package controllers

import (
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

// ListMods returns the list of mods for a server
func ListMods(c fiber.Ctx) error {
	id := c.Params("id")
	mods, err := services.ListMods(state, id)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}
	return utils.SuccessResponse(c, "OK", mods)
}

// UploadMod handles mod file upload
func UploadMod(c fiber.Ctx) error {
	id := c.Params("id")

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, "No file uploaded", fiber.StatusBadRequest)
	}

	filename := filepath.Base(file.Filename)

	src, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, "Failed to open uploaded file", fiber.StatusInternalServerError)
	}
	defer src.Close()

	if err := services.UploadMod(state, id, filename, src); err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	utils.LogAudit("admin", "MOD_UPLOAD", id, "Uploaded mod: "+filename)
	return utils.SuccessResponse(c, "Mod uploaded", nil)
}

// DeleteMod removes a mod file from the server
func DeleteMod(c fiber.Ctx) error {
	id := c.Params("id")
	name := c.Params("name")
	if err := services.DeleteMod(state, id, name); err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}
	utils.LogAudit("admin", "MOD_DELETE", id, "Deleted mod: "+name)
	return utils.SuccessResponse(c, "Deleted", nil)
}
