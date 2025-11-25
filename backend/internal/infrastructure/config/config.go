package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	General struct {
		RPS   float64 `json:"rps"`   // Requests per second
		Burst int     `json:"burst"` // Burst size
	} `json:"general"`
	Auth struct {
		RPM   int `json:"rpm"`   // Requests per minute
		Burst int `json:"burst"` // Burst size
	} `json:"auth"`
	Strict struct {
		RPM   int `json:"rpm"`   // Requests per minute
		Burst int `json:"burst"` // Burst size
	} `json:"strict"`
}

// AppConfig holds application configuration
type AppConfig struct {
	RateLimit RateLimitConfig `json:"rate_limit"`
}

var appConfig *AppConfig

// LoadConfig loads configuration from Vault or environment variables
func LoadConfig() (*AppConfig, error) {
	zapLog := logger.GetLogger()

	// Try to load from Vault first
	config, err := loadConfigFromVault()
	if err == nil && config != nil {
		zapLog.Info("Configuration loaded from Vault")
		appConfig = config
		return appConfig, nil
	}

	// Fallback to environment variables
	zapLog.Info("Loading configuration from environment variables")
	config = loadConfigFromEnv()
	appConfig = config
	return appConfig, nil
}

// loadConfigFromVault loads configuration from Vault
func loadConfigFromVault() (*AppConfig, error) {
	zapLog := logger.GetLogger()

	// Check if Vault is configured
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultPath := os.Getenv("VAULT_CONFIG_PATH")

	if vaultAddr == "" || vaultToken == "" {
		return nil, fmt.Errorf("Vault not configured")
	}

	if vaultPath == "" {
		vaultPath = "secret/data/dms-app" // Default path
	}

	// Create Vault client directly
	client, err := api.NewClient(&api.Config{Address: vaultAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}
	client.SetToken(vaultToken)
	
	// Read secret from Vault
	secret, err := client.Logical().Read(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from Vault: %w", err)
	}
	
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", vaultPath)
	}
	
	// Handle KV v2 format
	var data map[string]interface{}
	if secret.Data["data"] != nil {
		data = secret.Data["data"].(map[string]interface{})
	} else {
		data = secret.Data
	}

	// Try to extract rate_limit config
	// Support multiple formats:
	// 1. rate_limit as JSON string
	// 2. rate_limit_config as JSON string
	// 3. rate_limit as map
	// 4. Direct rate_limit structure in data
	
	var rateLimitData interface{}
	var found bool
	
	// Try different possible keys
	possibleKeys := []string{"rate_limit", "rate_limit_config", "app_config"}
	for _, key := range possibleKeys {
		if val, ok := data[key]; ok {
			rateLimitData = val
			found = true
			break
		}
	}
	
	if !found {
		// If no rate_limit key found, check if data itself is the config
		if _, ok := data["general"]; ok {
			rateLimitData = data
			found = true
		}
	}
	
	if !found {
		return nil, fmt.Errorf("rate_limit config not found in Vault secret")
	}
	
	// Parse rate_limit data
	config := &AppConfig{}
	
	switch v := rateLimitData.(type) {
	case string:
		// JSON string - parse it
		var rateLimitMap map[string]interface{}
		if err := json.Unmarshal([]byte(v), &rateLimitMap); err != nil {
			// Try parsing as full config
			if err := json.Unmarshal([]byte(v), config); err != nil {
				return nil, fmt.Errorf("failed to parse rate_limit JSON: %w", err)
			}
		} else {
			// Extract rate_limit from map
			if rl, ok := rateLimitMap["rate_limit"].(map[string]interface{}); ok {
				parseRateLimitConfig(rl, config)
			} else {
				// Assume the map itself is rate_limit config
				parseRateLimitConfig(rateLimitMap, config)
			}
		}
	case map[string]interface{}:
		// Map structure
		if rl, ok := v["rate_limit"].(map[string]interface{}); ok {
			parseRateLimitConfig(rl, config)
		} else {
			// Assume the map itself is rate_limit config
			parseRateLimitConfig(v, config)
		}
	default:
		return nil, fmt.Errorf("unexpected rate_limit type: %T", rateLimitData)
	}
	
	zapLog.Info("Configuration loaded from Vault",
		zap.String("path", vaultPath),
		zap.Float64("general_rps", config.RateLimit.General.RPS),
		zap.Int("auth_rpm", config.RateLimit.Auth.RPM),
	)
	
	return config, nil
}

// parseRateLimitConfig parses rate limit configuration from map
func parseRateLimitConfig(rateLimitMap map[string]interface{}, config *AppConfig) {
	// Parse general
	if general, ok := rateLimitMap["general"].(map[string]interface{}); ok {
		if rps, ok := general["rps"].(float64); ok {
			config.RateLimit.General.RPS = rps
		}
		if burst, ok := general["burst"].(float64); ok {
			config.RateLimit.General.Burst = int(burst)
		}
	}
	
	// Parse auth
	if auth, ok := rateLimitMap["auth"].(map[string]interface{}); ok {
		if rpm, ok := auth["rpm"].(float64); ok {
			config.RateLimit.Auth.RPM = int(rpm)
		}
		if burst, ok := auth["burst"].(float64); ok {
			config.RateLimit.Auth.Burst = int(burst)
		}
	}
	
	// Parse strict
	if strict, ok := rateLimitMap["strict"].(map[string]interface{}); ok {
		if rpm, ok := strict["rpm"].(float64); ok {
			config.RateLimit.Strict.RPM = int(rpm)
		}
		if burst, ok := strict["burst"].(float64); ok {
			config.RateLimit.Strict.Burst = int(burst)
		}
	}
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() *AppConfig {
	config := &AppConfig{}

	// General rate limit
	generalRPS := getEnvFloat("RATE_LIMIT_GENERAL_RPS", 500.0)
	generalBurst := getEnvInt("RATE_LIMIT_GENERAL_BURST", 500)
	config.RateLimit.General.RPS = generalRPS
	config.RateLimit.General.Burst = generalBurst

	// Auth rate limit
	authRPM := getEnvInt("RATE_LIMIT_AUTH_RPM", 5)
	authBurst := getEnvInt("RATE_LIMIT_AUTH_BURST", 5)
	config.RateLimit.Auth.RPM = authRPM
	config.RateLimit.Auth.Burst = authBurst

	// Strict rate limit
	strictRPM := getEnvInt("RATE_LIMIT_STRICT_RPM", 50)
	strictBurst := getEnvInt("RATE_LIMIT_STRICT_BURST", 50)
	config.RateLimit.Strict.RPM = strictRPM
	config.RateLimit.Strict.Burst = strictBurst

	return config
}

// GetConfig returns the loaded configuration
func GetConfig() *AppConfig {
	if appConfig == nil {
		// Load default config if not loaded
		appConfig = loadConfigFromEnv()
	}
	return appConfig
}

// GetGeneralRateLimit returns rate limit for general API
func (c *RateLimitConfig) GetGeneralRateLimit() (rate.Limit, int) {
	return rate.Limit(c.General.RPS), c.General.Burst
}

// GetAuthRateLimit returns rate limit for auth endpoints
func (c *RateLimitConfig) GetAuthRateLimit() (rate.Limit, int) {
	// Convert RPM to rate.Limit
	limit := rate.Every(time.Minute / time.Duration(c.Auth.RPM))
	return limit, c.Auth.Burst
}

// GetStrictRateLimit returns rate limit for strict endpoints
func (c *RateLimitConfig) GetStrictRateLimit() (rate.Limit, int) {
	// Convert RPM to rate.Limit
	limit := rate.Every(time.Minute / time.Duration(c.Strict.RPM))
	return limit, c.Strict.Burst
}

// Helper functions
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

