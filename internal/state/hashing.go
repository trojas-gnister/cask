// Package state manages config state persistence and change detection.
package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// HashString returns the SHA256 hex digest of a string.
func HashString(value string) string {
	h := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", h)
}

// HashData returns the SHA256 hex digest of any JSON-serializable value.
// Uses json.Marshal which sorts map keys deterministically in Go.
func HashData(data any) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("hashing: marshal failed: %w", err)
	}
	return HashString(string(b)), nil
}
