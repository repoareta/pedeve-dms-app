package http

import (
	"os/exec"
	"path/filepath"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RegenerateSwagger regenerate dokumentasi Swagger
// @Summary      Regenerate Swagger Documentation
// @Description  Regenerate Swagger documentation dari source code. Endpoint ini berguna untuk development agar Swagger selalu up-to-date tanpa perlu restart server.
// @Tags         System
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/swagger/regenerate [post]
func RegenerateSwagger(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()

	// Ambil working directory (folder backend)
	workDir := filepath.Join(".")
	
	// Jalankan command swag init
	cmd := exec.Command("go", "run", "github.com/swaggo/swag/cmd/swag@latest", "init", "-g", "cmd/api/main.go", "-o", "docs")
	cmd.Dir = workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		zapLog.Error("Failed to regenerate Swagger",
			zap.Error(err),
			zap.String("output", string(output)),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "regeneration_failed",
			"message": "Failed to regenerate Swagger documentation",
			"details": string(output),
		})
	}

	zapLog.Info("Swagger documentation regenerated successfully",
		zap.String("output", string(output)),
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Swagger documentation regenerated successfully",
		"output":  string(output),
	})
}

// GetSwaggerJSON serves the swagger.json file with auto-reload headers
func GetSwaggerJSON(c *fiber.Ctx) error {
	// Set header untuk prevent caching dan enable auto-reload
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	
	// Serve file swagger.json dari folder docs
	return c.SendFile("docs/swagger.json")
}

// GetSwaggerYAML serves the swagger.yaml file with auto-reload headers
func GetSwaggerYAML(c *fiber.Ctx) error {
	// Set header untuk prevent caching dan enable auto-reload
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	
	// Serve file swagger.yaml dari folder docs
	return c.SendFile("docs/swagger.yaml")
}

