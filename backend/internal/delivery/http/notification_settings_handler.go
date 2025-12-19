package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
)

// NotificationSettingsHandler handles notification settings HTTP requests
type NotificationSettingsHandler struct {
	settingsUC usecase.NotificationSettingsUseCase
}

// NewNotificationSettingsHandler creates a new notification settings handler
func NewNotificationSettingsHandler(settingsUC usecase.NotificationSettingsUseCase) *NotificationSettingsHandler {
	return &NotificationSettingsHandler{
		settingsUC: settingsUC,
	}
}

// GetSettings godoc
// @Summary      Get notification settings
// @Description  Get notification settings for the authenticated user
// @Tags         Notification Settings
// @Accept       json
// @Produce      json
// @Success      200  {object}  domain.NotificationSettingsModel
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /notification-settings [get]
// @Security     BearerAuth
func (h *NotificationSettingsHandler) GetSettings(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)

	settings, err := h.settingsUC.GetSettings(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get notification settings: " + err.Error(),
		})
	}

	return c.JSON(settings)
}

// UpdateSettings godoc
// @Summary      Update notification settings
// @Description  Update notification settings for the authenticated user
// @Tags         Notification Settings
// @Accept       json
// @Produce      json
// @Param        settings  body      object  true  "Notification settings (all fields optional)"
// @Success      200       {object}  domain.NotificationSettingsModel
// @Failure      400       {object}  domain.ErrorResponse
// @Failure      401       {object}  domain.ErrorResponse
// @Router       /notification-settings [put]
// @Security     BearerAuth
func (h *NotificationSettingsHandler) UpdateSettings(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)

	var req struct {
		EmailEnabled        *bool `json:"email_enabled"`
		InAppEnabled        *bool `json:"in_app_enabled"`
		ExpiryThresholdDays *int  `json:"expiry_threshold_days"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	settings, err := h.settingsUC.UpdateSettings(userID, req.EmailEnabled, req.InAppEnabled, req.ExpiryThresholdDays)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "expiry_threshold_days must be between 1 and 365 days" {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(settings)
}
