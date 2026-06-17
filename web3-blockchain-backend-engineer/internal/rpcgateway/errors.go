package rpcgateway

import "errors"

// Common errors for RPC gateway operations
var (
	// ErrNoAvailableProvider indicates no provider is available for the request
	ErrNoAvailableProvider = errors.New("no available RPC provider")

	// ErrAllProvidersFailed indicates all providers have failed
	ErrAllProvidersFailed = errors.New("all RPC providers have failed")

	// ErrProviderNotFound indicates the requested provider was not found
	ErrProviderNotFound = errors.New("RPC provider not found")

	// ErrInvalidChainID indicates an invalid chain ID was provided
	ErrInvalidChainID = errors.New("invalid chain ID")

	// ErrCircuitOpen indicates the circuit breaker is open for a provider
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrRequestTimeout indicates the RPC request timed out
	ErrRequestTimeout = errors.New("RPC request timeout")

	// ErrInvalidBlockRange indicates an invalid block range in logs filter
	ErrInvalidBlockRange = errors.New("invalid block range")

	// ErrBatchSizeTooLarge indicates batch size exceeds maximum allowed
	ErrBatchSizeTooLarge = errors.New("batch size too large")

	// ErrTransactionNotFound indicates the transaction was not found
	ErrTransactionNotFound = errors.New("transaction not found")

	// ErrReceiptNotFound indicates the receipt was not found
	ErrReceiptNotFound = errors.New("receipt not found")

	// ErrBlockNotFound indicates the block was not found
	ErrBlockNotFound = errors.New("block not found")
)

// ProviderError represents an error from a specific provider
type ProviderError struct {
	ProviderID int64
	ProviderName string
	Err        error
}

func (e *ProviderError) Error() string {
	if e.ProviderName != "" {
		return e.ProviderName + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError
func NewProviderError(providerID int64, providerName string, err error) *ProviderError {
	return &ProviderError{
		ProviderID:   providerID,
		ProviderName: providerName,
		Err:          err,
	}
}

// RPCError represents an RPC-level error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// RetryableError indicates the error may succeed on retry
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	var retryable *RetryableError
	if errors.As(err, &retryable) {
		return true
	}
	return false
}
