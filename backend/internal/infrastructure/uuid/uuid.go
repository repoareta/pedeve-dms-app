package uuid

import "github.com/google/uuid"

// GenerateUUID menghasilkan string UUID baru
func GenerateUUID() string {
	return uuid.New().String()
}

