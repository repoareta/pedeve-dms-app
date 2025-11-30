package http

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/sonarqube"
	"go.uber.org/zap"
)

type SonarQubeHandler struct {
	logger *zap.Logger
}

func NewSonarQubeHandler() *SonarQubeHandler {
	return &SonarQubeHandler{
		logger: logger.GetLogger(),
	}
}

// GetIssues fetches issues from SonarCloud
// @Summary      Get SonarCloud Issues
// @Description  Mengambil daftar issues dari SonarCloud untuk project ini (hanya superadmin/admin)
// @Tags         SonarQube
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        severities  query     []string  false  "Filter by severity (BLOCKER, CRITICAL, MAJOR, MINOR, INFO)"
// @Param        types      query     []string  false  "Filter by type (BUG, VULNERABILITY, CODE_SMELL)"
// @Param        statuses   query     []string  false  "Filter by status (OPEN, CONFIRMED, REOPENED, RESOLVED)"
// @Success      200        {object}  sonarqube.IssuesResponse
// @Failure      403        {object}  domain.ErrorResponse
// @Failure      500        {object}  domain.ErrorResponse
// @Router       /sonarqube/issues [get]
func (h *SonarQubeHandler) GetIssues(c *fiber.Ctx) error {
	// Check if user is superadmin or admin
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		h.logger.Warn("RoleName not found in context", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		h.logger.Warn("Invalid roleName type in context", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	// Only superadmin and admin can access
	if roleName != "superadmin" && roleName != "admin" {
		h.logger.Warn("Unauthorized access attempt to SonarCloud issues",
			zap.String("role", roleName),
			zap.String("path", c.Path()),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin and admin can access SonarCloud issues",
		})
	}

	// Create SonarCloud client
	client, err := sonarqube.NewSonarCloudClient()
	if err != nil {
		h.logger.Error("Failed to create SonarCloud client", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "configuration_error",
			Message: fmt.Sprintf("Failed to initialize SonarCloud client: %v. Please check SONARCLOUD_URL, SONARCLOUD_TOKEN, and SONARCLOUD_PROJECT_KEY environment variables.", err),
		})
	}

	// Get query parameters (support multiple values with same name)
	// Frontend sends: severities=BLOCKER&severities=CRITICAL&severities=MAJOR
	// Use QueryArgs() to get all values for each parameter
	queryArgs := c.Request().URI().QueryArgs()

	var severityList []string
	queryArgs.VisitAll(func(key, value []byte) {
		if string(key) == "severities" {
			severityList = append(severityList, string(value))
		}
	})

	var typeList []string
	queryArgs.VisitAll(func(key, value []byte) {
		if string(key) == "types" {
			typeList = append(typeList, string(value))
		}
	})

	var statusList []string
	queryArgs.VisitAll(func(key, value []byte) {
		if string(key) == "statuses" {
			statusList = append(statusList, string(value))
		}
	})

	// Fetch issues
	issuesResp, err := client.GetIssues(severityList, typeList, statusList)
	if err != nil {
		h.logger.Error("Failed to fetch issues from SonarCloud",
			zap.Error(err),
			zap.Strings("severities", severityList),
			zap.Strings("types", typeList),
			zap.Strings("statuses", statusList),
		)

		// Check if error is about authentication
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "authentication") || strings.Contains(errorMsg, "Unauthorized") {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "authentication_error",
				Message: "SonarCloud authentication failed. Please check SONARCLOUD_TOKEN in Secret Manager (Vault) or environment variables. Make sure backend is restarted after storing secrets.",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "sonarqube_error",
			Message: fmt.Sprintf("Failed to fetch issues from SonarCloud: %v. Please check SonarCloud configuration and token.", err),
		})
	}

	// Log response for debugging
	h.logger.Info("SonarCloud issues fetched successfully",
		zap.Int("total", issuesResp.Total),
		zap.Int("issues_count", len(issuesResp.Issues)),
		zap.Int("components_count", len(issuesResp.Components)),
		zap.Strings("severities", severityList),
		zap.Strings("types", typeList),
		zap.Strings("statuses", statusList),
	)

	// Return response even if total is 0 (no issues match the filter)
	return c.JSON(issuesResp)
}

// ExportIssues exports issues as JSON file
// @Summary      Export SonarCloud Issues
// @Description  Export daftar issues dari SonarCloud dalam format JSON untuk download (hanya superadmin/admin)
// @Tags         SonarQube
// @Accept       json
// @Produce      application/json
// @Security     BearerAuth
// @Param        severities  query     []string  false  "Filter by severity (BLOCKER, CRITICAL, MAJOR, MINOR, INFO)"
// @Param        types      query     []string  false  "Filter by type (BUG, VULNERABILITY, CODE_SMELL)"
// @Param        statuses   query     []string  false  "Filter by status (OPEN, CONFIRMED, REOPENED, RESOLVED)"
// @Success      200        {file}    application/json
// @Failure      403        {object}  domain.ErrorResponse
// @Failure      500        {object}  domain.ErrorResponse
// @Router       /sonarqube/issues/export [get]
func (h *SonarQubeHandler) ExportIssues(c *fiber.Ctx) error {
	// Check if user is superadmin or admin
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		h.logger.Warn("RoleName not found in context", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		h.logger.Warn("Invalid roleName type in context", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	// Only superadmin and admin can access
	if roleName != "superadmin" && roleName != "admin" {
		h.logger.Warn("Unauthorized access attempt to export SonarCloud issues",
			zap.String("role", roleName),
			zap.String("path", c.Path()),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin and admin can export SonarCloud issues",
		})
	}

	// Create SonarCloud client
	client, err := sonarqube.NewSonarCloudClient()
	if err != nil {
		h.logger.Error("Failed to create SonarCloud client", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "configuration_error",
			Message: fmt.Sprintf("Failed to initialize SonarCloud client: %v. Please check SONARCLOUD_URL, SONARCLOUD_TOKEN, and SONARCLOUD_PROJECT_KEY environment variables.", err),
		})
	}

	// Get query parameters (support multiple values with same name)
	// Frontend sends: severities=BLOCKER&severities=CRITICAL&severities=MAJOR
	// Use QueryArgs() to get all values for each parameter
	queryArgs := c.Request().URI().QueryArgs()

	var severityList []string
	queryArgs.VisitAll(func(key, value []byte) {
		if string(key) == "severities" {
			severityList = append(severityList, string(value))
		}
	})

	var typeList []string
	queryArgs.VisitAll(func(key, value []byte) {
		if string(key) == "types" {
			typeList = append(typeList, string(value))
		}
	})

	var statusList []string
	queryArgs.VisitAll(func(key, value []byte) {
		if string(key) == "statuses" {
			statusList = append(statusList, string(value))
		}
	})

	// Fetch issues
	issuesResp, err := client.GetIssues(severityList, typeList, statusList)
	if err != nil {
		h.logger.Error("Failed to fetch issues from SonarCloud",
			zap.Error(err),
			zap.Strings("severities", severityList),
			zap.Strings("types", typeList),
			zap.Strings("statuses", statusList),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "sonarqube_error",
			Message: fmt.Sprintf("Failed to fetch issues from SonarCloud: %v. Please check SonarCloud configuration and token.", err),
		})
	}

	// Set headers for file download
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", "attachment; filename=sonarqube-issues.json")

	// Return JSON response (Fiber will automatically serialize)
	return c.JSON(issuesResp)
}

// GetSoftwareQualityMetrics fetches Software Quality metrics from SonarCloud
// @Summary      Get Software Quality Metrics
// @Description  Mengambil metrics Software Quality (Security, Reliability, Maintainability) dari SonarCloud (hanya superadmin/admin)
// @Tags         SonarQube
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  sonarqube.SoftwareQualityMetrics
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /sonarqube/quality [get]
func (h *SonarQubeHandler) GetSoftwareQualityMetrics(c *fiber.Ctx) error {
	// Check if user is superadmin or admin
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		h.logger.Warn("RoleName not found in context", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		h.logger.Warn("RoleName is not a string", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user role",
		})
	}

	// Only superadmin and admin can access
	if roleName != "superadmin" && roleName != "admin" {
		h.logger.Warn("Unauthorized access attempt to SonarQube metrics",
			zap.String("role", roleName),
			zap.String("path", c.Path()),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin and admin can access SonarQube metrics",
		})
	}

	// Create SonarCloud client
	client, err := sonarqube.NewSonarCloudClient()
	if err != nil {
		h.logger.Error("Failed to create SonarCloud client",
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "configuration_error",
			Message: fmt.Sprintf("Failed to initialize SonarCloud client: %v", err),
		})
	}

	// Fetch Software Quality metrics
	metrics, err := client.GetSoftwareQualityMetrics()
	if err != nil {
		h.logger.Error("Failed to fetch Software Quality metrics from SonarCloud",
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "api_error",
			Message: fmt.Sprintf("Failed to fetch Software Quality metrics: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(metrics)
}

// GetStatus returns the status of SonarQube Monitor feature
// @Summary      Get SonarQube Monitor Status
// @Description  Mengecek apakah fitur SonarQube Monitor tersedia/enabled (public endpoint, tidak perlu auth)
// @Tags         SonarQube
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]bool
// @Router       /sonarqube/status [get]
func (h *SonarQubeHandler) GetStatus(c *fiber.Ctx) error {
	// Check if SonarQube Monitor is enabled by checking environment variable
	// This endpoint is always available, but returns enabled=false if feature is disabled
	enableSonarQubeMonitor := os.Getenv("ENABLE_SONARQUBE_MONITOR")
	env := os.Getenv("ENV")

	enabled := true
	if enableSonarQubeMonitor != "" {
		enabled = enableSonarQubeMonitor == "true"
	} else if env == "production" {
		// Default: disable di production
		enabled = false
	}

	// Also check if client can be created (secrets available)
	if enabled {
		_, err := sonarqube.NewSonarCloudClient()
		if err != nil {
			enabled = false
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"enabled": enabled,
	})
}
