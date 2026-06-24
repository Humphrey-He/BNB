package handlers

import (
	"database/sql"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func parseOptionalInt64(c *gin.Context, key string) (int64, bool, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, false, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return parsed, true, nil
}

func parseLimit(c *gin.Context, defaultLimit, maxLimit int) (int, error) {
	raw := strings.TrimSpace(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if limit <= 0 {
		return 0, errors.New("limit must be positive")
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit, nil
}

func respondRepositoryError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	handlerLogger().Error("handler repository error", "error", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parsePositiveIntegerString(raw string) (*big.Int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, errors.New("value is required")
	}
	n, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, errors.New("value must be a base-10 integer")
	}
	if n.Sign() <= 0 {
		return nil, errors.New("value must be positive")
	}
	return n, nil
}
