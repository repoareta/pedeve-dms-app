package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertNoDuplicates checks that a slice has no duplicate IDs
func AssertNoDuplicates(t *testing.T, items []interface{}, getID func(interface{}) string, itemType string) {
	ids := make(map[string]int)
	for _, item := range items {
		id := getID(item)
		ids[id]++
	}

	for id, count := range ids {
		assert.Equal(t, 1, count, "%s with ID %s should appear only once, but appears %d times", itemType, id, count)
	}
}

// RequireNoDuplicates is like AssertNoDuplicates but uses require (stops test on failure)
func RequireNoDuplicates(t *testing.T, items []interface{}, getID func(interface{}) string, itemType string) {
	ids := make(map[string]int)
	for _, item := range items {
		id := getID(item)
		ids[id]++
	}

	for id, count := range ids {
		require.Equal(t, 1, count, "%s with ID %s should appear only once, but appears %d times", itemType, id, count)
	}
}

