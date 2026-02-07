package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// parseOptionalTime parses an optional RFC3339 query parameter.
// Returns nil if the param is empty, or responds with an error and returns false.
func parseOptionalTime(c *gin.Context, param string) (*time.Time, bool) {
	raw := c.Query(param)
	if raw == "" {
		return nil, true
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		respondError(c, http.StatusBadRequest, param+" tidak valid")
		return nil, false
	}
	return &parsed, true
}

func parsePageAndLimit(
	c *gin.Context,
	defaultLimit int,
	maxLimit int,
) (int, int) {
	page := parsePositiveIntQuery(c, "page", 1)
	limit := parsePositiveIntQuery(c, "limit", defaultLimit)
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

func parsePositiveIntQuery(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
