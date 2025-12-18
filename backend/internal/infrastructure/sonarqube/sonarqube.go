package sonarqube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/secrets"
	"go.uber.org/zap"
)

// SonarCloudClient handles communication with SonarCloud API
type SonarCloudClient struct {
	baseURL   string
	token     string
	projectKey string
	httpClient *http.Client
}

// Issue represents a SonarCloud issue
type Issue struct {
	Key       string `json:"key"`
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	Component string `json:"component"`
	Project   string `json:"project"`
	Line      int    `json:"line"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Effort    string `json:"effort"`
	Debt      string `json:"debt"`
	Author    string `json:"author"`
	CreationDate string `json:"creationDate"`
	UpdateDate   string `json:"updateDate"`
}

// IssuesResponse represents the response from SonarCloud issues/search API
type IssuesResponse struct {
	Total    int     `json:"total"`
	P        int     `json:"p"`
	Ps       int     `json:"ps"`
	Paging   Paging  `json:"paging"`
	Issues   []Issue `json:"issues"`
	Components []Component `json:"components"`
}

// Paging represents pagination info
type Paging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

// Component represents a component in SonarCloud
type Component struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Qualifier string `json:"qualifier"`
	Path     string `json:"path,omitempty"`
}

// Measure represents a metric measure from SonarCloud
type Measure struct {
	Metric    string `json:"metric"`
	Value     string `json:"value"`
	BestValue bool   `json:"bestValue"`
}

// ComponentMeasuresResponse represents the response from SonarCloud measures/component API
type ComponentMeasuresResponse struct {
	Component struct {
		ID       string    `json:"id"`
		Key      string    `json:"key"`
		Name     string    `json:"name"`
		Qualifier string   `json:"qualifier"`
		Measures []Measure `json:"measures"`
	} `json:"component"`
}

// SoftwareQualityMetrics represents Software Quality metrics
type SoftwareQualityMetrics struct {
	Security      int `json:"security"`      // vulnerabilities count
	Reliability   int `json:"reliability"`  // bugs count
	Maintainability int `json:"maintainability"` // code_smells count
}

// NewSonarCloudClient creates a new SonarCloud client
// Supports reading from Secret Manager (Vault/GCP) or environment variables
func NewSonarCloudClient() (*SonarCloudClient, error) {
	zapLog := logger.GetLogger()

	// Get SONARCLOUD_URL
	baseURL, err := secrets.GetSecretWithFallback("SONARCLOUD_URL", "SONARCLOUD_URL", "https://sonarcloud.io")
	if err != nil {
		zapLog.Warn("Failed to get SONARCLOUD_URL from secret manager, using default",
			zap.Error(err),
		)
		baseURL = "https://sonarcloud.io"
	}

	// Get SONARCLOUD_TOKEN (required)
	token, err := secrets.GetSecretWithFallback("SONARCLOUD_TOKEN", "SONARCLOUD_TOKEN", "")
	if err != nil || token == "" {
		// Try direct environment variable as last resort
		token = os.Getenv("SONARCLOUD_TOKEN")
		if token == "" {
			zapLog.Error("SONARCLOUD_TOKEN not found",
				zap.Error(err),
				zap.String("source", "secret_manager_and_env"),
			)
			return nil, fmt.Errorf("SONARCLOUD_TOKEN is required but not found in Secret Manager or environment variables")
		}
		zapLog.Info("Using SONARCLOUD_TOKEN from environment variable")
	} else {
		zapLog.Info("Using SONARCLOUD_TOKEN from Secret Manager")
	}

	// Get SONARCLOUD_PROJECT_KEY (required)
	projectKey, err := secrets.GetSecretWithFallback("SONARCLOUD_PROJECT_KEY", "SONARCLOUD_PROJECT_KEY", "")
	if err != nil || projectKey == "" {
		// Try direct environment variable as last resort
		projectKey = os.Getenv("SONARCLOUD_PROJECT_KEY")
		if projectKey == "" {
			zapLog.Error("SONARCLOUD_PROJECT_KEY not found",
				zap.Error(err),
				zap.String("source", "secret_manager_and_env"),
			)
			return nil, fmt.Errorf("SONARCLOUD_PROJECT_KEY is required but not found in Secret Manager or environment variables")
		}
		zapLog.Info("Using SONARCLOUD_PROJECT_KEY from environment variable")
	} else {
		zapLog.Info("Using SONARCLOUD_PROJECT_KEY from Secret Manager")
	}

	zapLog.Info("SonarCloud client initialized",
		zap.String("base_url", baseURL),
		zap.String("project_key", projectKey),
		zap.Bool("token_set", token != ""),
	)

	return &SonarCloudClient{
		baseURL:    baseURL,
		token:      token,
		projectKey: projectKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetIssues fetches issues from SonarCloud
// Filters: severity (BLOCKER, CRITICAL, MAJOR), type (BUG, VULNERABILITY), status (OPEN, CONFIRMED, REOPENED)
func (c *SonarCloudClient) GetIssues(severities []string, types []string, statuses []string) (*IssuesResponse, error) {
	zapLog := logger.GetLogger()

	// Build API endpoint
	apiURL := fmt.Sprintf("%s/api/issues/search", c.baseURL)

	// Build query parameters
	params := url.Values{}
	params.Set("componentKeys", c.projectKey)
	params.Set("resolved", "false") // Only get unresolved issues

	// Add severity filter (for VAPT: BLOCKER, CRITICAL, MAJOR)
	// Note: When combining multiple filters, SonarCloud API may return 0 results
	// if the combination is too restrictive. Consider using fewer filters.
	if len(severities) > 0 {
		for _, severity := range severities {
			params.Add("severities", severity)
		}
	}
	// No default severity filter - let API return all severities if not specified

	// Add type filter (for VAPT: BUG, VULNERABILITY, CODE_SMELL)
	if len(types) > 0 {
		for _, issueType := range types {
			params.Add("types", issueType)
		}
	}
	// No default type filter - let API return all types if not specified

	// Add status filter (OPEN, CONFIRMED, REOPENED)
	if len(statuses) > 0 {
		for _, status := range statuses {
			params.Add("statuses", status)
		}
	} else {
		// Default: Only OPEN status (most common for unresolved issues)
		// Using only OPEN avoids conflicts with multiple filter combinations
		params.Add("statuses", "OPEN")
	}

	// Set pagination
	params.Set("p", "1")
	params.Set("ps", "500") // Max 500 issues per request

	// Build full URL
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header (Basic Auth with token)
	// SonarCloud expects token as username with empty password
	if c.token == "" {
		return nil, fmt.Errorf("SonarCloud token is empty")
	}
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:", c.token)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	
	// Log for debugging (don't log full token)
	zapLog.Debug("SonarCloud authentication",
		zap.String("token_length", fmt.Sprintf("%d", len(c.token))),
		zap.String("token_prefix", func() string {
			if len(c.token) > 10 {
				return c.token[:10] + "..."
			}
			return "***"
		}()),
	)

	// Make request
	zapLog.Info("Fetching issues from SonarCloud",
		zap.String("url", apiURL),
		zap.String("project", c.projectKey),
		zap.String("full_url", fullURL),
		zap.Strings("severities", severities),
		zap.Strings("types", types),
		zap.Strings("statuses", statuses),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		zapLog.Error("SonarCloud API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
			zap.String("url", fullURL),
			zap.String("token_set", func() string {
				if c.token == "" {
					return "false"
				}
				return fmt.Sprintf("true (length: %d)", len(c.token))
			}()),
		)
		
		// Provide more helpful error message for 401
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("SonarCloud authentication failed (401): Invalid or expired token. Please check SONARCLOUD_TOKEN in Secret Manager or environment variables")
		}
		
		return nil, fmt.Errorf("SonarCloud API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var issuesResp IssuesResponse
	if err := json.Unmarshal(body, &issuesResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	zapLog.Info("Successfully fetched issues from SonarCloud",
		zap.Int("total", issuesResp.Total),
		zap.Int("count", len(issuesResp.Issues)),
		zap.String("url", fullURL),
		zap.Strings("severities", severities),
		zap.Strings("types", types),
		zap.Strings("statuses", statuses),
	)
	
	// Log warning if no issues found with filters
	if issuesResp.Total == 0 && (len(severities) > 0 || len(types) > 0 || len(statuses) > 0) {
		zapLog.Warn("No issues found with current filters. Try removing some filters to see more issues.",
			zap.Strings("severities", severities),
			zap.Strings("types", types),
			zap.Strings("statuses", statuses),
		)
	}

	return &issuesResp, nil
}

// GetSoftwareQualityMetrics fetches Software Quality metrics from SonarCloud
// Returns Security (vulnerabilities), Reliability (bugs), and Maintainability (code_smells) counts
func (c *SonarCloudClient) GetSoftwareQualityMetrics() (*SoftwareQualityMetrics, error) {
	zapLog := logger.GetLogger()

	// Build API endpoint
	apiURL := fmt.Sprintf("%s/api/measures/component", c.baseURL)

	// Build query parameters
	params := url.Values{}
	params.Set("component", c.projectKey)
	params.Set("metricKeys", "vulnerabilities,bugs,code_smells")

	// Build full URL
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:", c.token)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))

	zapLog.Info("Fetching Software Quality metrics from SonarCloud",
		zap.String("url", apiURL),
		zap.String("project", c.projectKey),
		zap.String("full_url", fullURL),
	)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		zapLog.Error("SonarCloud API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
			zap.String("url", fullURL),
		)
		return nil, fmt.Errorf("SonarCloud API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var measuresResp ComponentMeasuresResponse
	if err := json.Unmarshal(body, &measuresResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract metrics
	metrics := &SoftwareQualityMetrics{
		Security:       0,
		Reliability:    0,
		Maintainability: 0,
	}

	for _, measure := range measuresResp.Component.Measures {
		switch measure.Metric {
		case "vulnerabilities":
			var count int
			if _, err := fmt.Sscanf(measure.Value, "%d", &count); err == nil {
				metrics.Security = count
			}
		case "bugs":
			var count int
			if _, err := fmt.Sscanf(measure.Value, "%d", &count); err == nil {
				metrics.Reliability = count
			}
		case "code_smells":
			var count int
			if _, err := fmt.Sscanf(measure.Value, "%d", &count); err == nil {
				metrics.Maintainability = count
			}
		}
	}

	zapLog.Info("Successfully fetched Software Quality metrics from SonarCloud",
		zap.Int("security", metrics.Security),
		zap.Int("reliability", metrics.Reliability),
		zap.Int("maintainability", metrics.Maintainability),
	)

	return metrics, nil
}

