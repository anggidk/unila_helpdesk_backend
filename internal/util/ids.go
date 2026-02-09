package util

import (
	"strings"

	"github.com/google/uuid"
)

func NewID(length int) string {
	if length <= 0 {
		return ""
	}
	raw := strings.ReplaceAll(uuid.NewString(), "-", "")
	if length >= len(raw) {
		return raw
	}
	return raw[:length]
}
