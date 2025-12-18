package audit

import (
	"os"
	"strings"
	"sync"
)

var (
	logViewsOnce sync.Once
	// Default dimatikan supaya tidak membanjiri log/storage.
	// Bisa dihidupkan dengan AUDIT_LOG_VIEWS=true.
	logViews = false
)

// ShouldLogView mengembalikan apakah aktivitas "view/list" perlu dicatat ke audit log.
// Bisa diatur lewat env AUDIT_LOG_VIEWS (default: true). Set ke "false" untuk mematikan.
func ShouldLogView() bool {
	logViewsOnce.Do(func() {
		val := strings.ToLower(strings.TrimSpace(os.Getenv("AUDIT_LOG_VIEWS")))
		if val == "true" || val == "1" || val == "yes" || val == "on" {
			logViews = true
		}
	})
	return logViews
}
