package http

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	notificationUC usecase.NotificationUseCase
}

func NewNotificationHandler(notificationUC usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		notificationUC: notificationUC,
	}
}

// GetNotifications godoc
// @Summary      Get user notifications
// @Description  Get list of notifications for the authenticated user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        unread_only  query     bool    false  "Filter unread only"
// @Param        limit        query     int     false  "Limit results (default: 5)"
// @Success      200          {object}  map[string]interface{}
// @Failure      401          {object}  domain.ErrorResponse
// @Router       /notifications [get]
// @Security     BearerAuth
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	
	// Get role and company for RBAC
	roleNameVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	
	roleName := ""
	if roleNameVal != nil {
		if rn, ok := roleNameVal.(string); ok {
			roleName = rn
		}
	}
	
	var companyID *string
	if companyIDVal != nil {
		if cid, ok := companyIDVal.(*string); ok && cid != nil {
			companyID = cid
		} else if cidStr, ok := companyIDVal.(string); ok && cidStr != "" {
			companyID = &cidStr
		}
	}
	
	unreadOnly := false
	if c.Query("unread_only") == "true" {
		unreadOnly = true
	}
	
	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	// Use RBAC-aware method untuk konsistensi dengan GetUnreadCount
	var unreadOnlyPtr *bool
	if unreadOnly {
		unreadOnlyPtr = &unreadOnly
	}
	
	zapLog := logger.GetLogger()
	zapLog.Info("GetNotifications called",
		zap.String("user_id", userID),
		zap.String("role_name", roleName),
		zap.Any("company_id", companyID),
		zap.Bool("unread_only", unreadOnly),
		zap.Int("limit", limit),
	)
	
	notifications, total, _, err := h.notificationUC.GetNotificationsWithRBAC(
		userID, roleName, companyID, unreadOnlyPtr, nil, 1, limit,
	)
	if err != nil {
		// Log error untuk debugging
		zapLog.Error("Failed to fetch notifications", zap.Error(err), zap.String("user_id", userID))
		
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: fmt.Sprintf("Failed to fetch notifications: %v", err),
		})
	}
	
	zapLog.Info("GetNotifications success",
		zap.String("user_id", userID),
		zap.String("role_name", roleName),
		zap.Int("count", len(notifications)),
		zap.Int64("total", total),
	)
	
	return c.JSON(fiber.Map{
		"data": notifications,
	})
}

// GetNotificationsWithFilters godoc
// @Summary      Get user notifications with filters
// @Description  Get paginated list of notifications with filters (for inbox page)
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        unread_only        query     bool    false  "Filter unread only"
// @Param        days_until_expiry  query     int     false  "Filter by days until expiry (3, 7, 30)"
// @Param        page               query     int     false  "Page number (default: 1)"
// @Param        page_size          query     int     false  "Page size (default: 10)"
// @Success      200                {object}  map[string]interface{}
// @Failure      401                {object}  domain.ErrorResponse
// @Router       /notifications/inbox [get]
// @Security     BearerAuth
func (h *NotificationHandler) GetNotificationsWithFilters(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	
	// Get role and company for RBAC
	roleNameVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	
	roleName := ""
	if roleNameVal != nil {
		if rn, ok := roleNameVal.(string); ok {
			roleName = rn
		}
	}
	
	var companyID *string
	if companyIDVal != nil {
		if cidPtr, ok := companyIDVal.(*string); ok && cidPtr != nil {
			companyID = cidPtr
		} else if cidStr, ok := companyIDVal.(string); ok {
			companyID = &cidStr
		}
	}
	
	var unreadOnly *bool
	if unreadStr := c.Query("unread_only"); unreadStr != "" {
		val := unreadStr == "true"
		unreadOnly = &val
	}
	
	var daysUntilExpiry *int
	if daysStr := c.Query("days_until_expiry"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			daysUntilExpiry = &days
		}
	}
	
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	
	pageSize := 10
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
		}
	}
	
	// Use RBAC-aware method
	notifications, total, totalPages, err := h.notificationUC.GetNotificationsWithRBAC(
		userID, roleName, companyID, unreadOnly, daysUntilExpiry, page, pageSize,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch notifications",
		})
	}
	
	return c.JSON(fiber.Map{
		"data":       notifications,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": totalPages,
	})
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Description  Mark a specific notification as read
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /notifications/:id/read [put]
// @Security     BearerAuth
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	notificationID := c.Params("id")
	
	// Get username for audit log
	username := ""
	if usernameVal := c.Locals("username"); usernameVal != nil {
		if u, ok := usernameVal.(string); ok {
			username = u
		}
	}
	
	if err := h.notificationUC.MarkAsRead(notificationID, userID); err != nil {
		if err.Error() == "notification not found" || err.Error() == "forbidden: notification does not belong to user" {
			// Audit log untuk failure
			if username != "" {
				audit.LogAction(userID, username, audit.ActionMarkNotificationRead, audit.ResourceNotification, notificationID, getClientIP(c), c.Get("User-Agent"), audit.StatusFailure, map[string]interface{}{
					"reason": "notification_not_found_or_forbidden",
				})
			}
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Notification not found",
			})
		}
		// Audit log untuk error
		if username != "" {
			audit.LogAction(userID, username, audit.ActionMarkNotificationRead, audit.ResourceNotification, notificationID, getClientIP(c), c.Get("User-Agent"), audit.StatusError, map[string]interface{}{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to mark notification as read",
		})
	}
	
	// Audit log untuk success
	if username != "" {
		audit.LogAction(userID, username, audit.ActionMarkNotificationRead, audit.ResourceNotification, notificationID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)
	}
	
	return c.JSON(fiber.Map{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead godoc
// @Summary      Mark all notifications as read
// @Description  Mark all notifications for the authenticated user as read
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /notifications/read-all [put]
// @Security     BearerAuth
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	
	if err := h.notificationUC.MarkAllAsRead(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to mark all notifications as read",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "All notifications marked as read",
	})
}

// GetUnreadCount godoc
// @Summary      Get unread notification count
// @Description  Get count of unread notifications for the authenticated user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /notifications/unread-count [get]
// @Security     BearerAuth
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	
	// Get role and company for RBAC
	roleNameVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	
	roleName := ""
	if roleNameVal != nil {
		if rn, ok := roleNameVal.(string); ok {
			roleName = rn
		}
	}
	
	var companyID *string
	if companyIDVal != nil {
		if cidPtr, ok := companyIDVal.(*string); ok && cidPtr != nil {
			companyID = cidPtr
		} else if cidStr, ok := companyIDVal.(string); ok {
			companyID = &cidStr
		}
	}
	
	// Use RBAC-aware method
	count, err := h.notificationUC.GetUnreadCountWithRBAC(userID, roleName, companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get unread count",
		})
	}
	
	return c.JSON(fiber.Map{
		"count": count,
	})
}

// StreamNotifications godoc
// @Summary      Stream notifications (SSE) - Placeholder
// @Description  Server-Sent Events stream for real-time notifications (currently returns empty, use polling instead)
// @Tags         notifications
// @Accept       json
// @Produce      text/event-stream
// @Success      200  {string}  text/event-stream
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /notifications/stream [get]
// @Security     BearerAuth
// @Note         SSE implementation akan ditambahkan nanti. Untuk sekarang, frontend menggunakan polling.
func (h *NotificationHandler) StreamNotifications(c *fiber.Ctx) error {
	// TODO: Implement SSE dengan benar nanti
	// Untuk sekarang, frontend akan menggunakan polling sebagai fallback
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	
	if _, err := c.WriteString(": SSE endpoint - use polling instead\n\n"); err != nil {
		return err
	}
	return nil
}

// DeleteAllNotifications godoc
// @Summary      Delete all notifications
// @Description  Hapus semua notifikasi sesuai RBAC: superadmin hapus semua, admin hapus company+descendants, user hapus sendiri
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /notifications/delete-all [delete]
// @Security     BearerAuth
func (h *NotificationHandler) DeleteAllNotifications(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	
	// Get role and company for RBAC
	roleNameVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	
	roleName := ""
	if roleNameVal != nil {
		if rn, ok := roleNameVal.(string); ok {
			roleName = rn
		}
	}
	
	var companyID *string
	if companyIDVal != nil {
		if cidPtr, ok := companyIDVal.(*string); ok && cidPtr != nil {
			companyID = cidPtr
		} else if cidStr, ok := companyIDVal.(string); ok {
			companyID = &cidStr
		}
	}
	
	// Use RBAC-aware method
	if err := h.notificationUC.DeleteAllWithRBAC(userID, roleName, companyID); err != nil {
		zapLog := logger.GetLogger()
		zapLog.Error("Failed to delete all notifications", zap.Error(err), zap.String("user_id", userID), zap.String("role", roleName))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete all notifications",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "All notifications deleted successfully",
	})
}

