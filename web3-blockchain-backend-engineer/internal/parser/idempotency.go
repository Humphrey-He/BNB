package parser

import (
	"fmt"
)

// MakeIdempotencyKey generates a unique idempotency key for a chain event
// Format: chain_id:tx_hash:log_index
func MakeIdempotencyKey(chainID int64, txHash string, logIndex uint) string {
	return fmt.Sprintf("%d:%s:%d", chainID, txHash, logIndex)
}

// IdempotencyKeyParser provides utilities for parsing idempotency keys
type IdempotencyKeyParser struct{}

// NewIdempotencyKeyParser creates a new IdempotencyKeyParser
func NewIdempotencyKeyParser() *IdempotencyKeyParser {
	return &IdempotencyKeyParser{}
}

// ParseIdempotencyKey parses an idempotency key into its components
func (p *IdempotencyKeyParser) ParseIdempotencyKey(key string) (chainID int64, txHash string, logIndex uint, err error) {
	_, err = fmt.Sscanf(key, "%d:%s:%d", &chainID, &txHash, &logIndex)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to parse idempotency key: %w", err)
	}
	return chainID, txHash, logIndex, nil
}
