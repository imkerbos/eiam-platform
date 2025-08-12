package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateTradeID generate TradeID
func GenerateTradeID() string {
	return uuid.New().String()
}

// GenerateTradeIDString generate TradeID based on source
func GenerateTradeIDString(source string) string {
	timestamp := time.Now().Format("20060102150405")
	shortUUID := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s_%s", source, timestamp, shortUUID)
}

// ParseTradeIDSource parse source from TradeID
func ParseTradeIDSource(tradeID string) string {
	if len(tradeID) < 8 {
		return "unknown"
	}

	// If it's standard UUID format, return api
	if len(tradeID) == 36 && tradeID[8] == '-' && tradeID[13] == '-' {
		return "api"
	}

	// Parse TradeID in format source_timestamp_uuid
	for i, char := range tradeID {
		if char == '_' {
			return tradeID[:i]
		}
	}

	return "unknown"
}
