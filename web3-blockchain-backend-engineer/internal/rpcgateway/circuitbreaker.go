package rpcgateway

import (
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed means normal operation
	StateClosed State = iota
	// StateOpen means failing fast
	StateOpen
	// StateHalfOpen means testing recovery
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for RPC calls
type CircuitBreaker struct {
	failureThreshold int           // Number of failures before opening
	cooldownPeriod   time.Duration // Cool-down time before half-open
	state            State         // Current state (closed/open/half-open)
	mu               sync.RWMutex
	failures         int           // Consecutive failure count
	lastFailure      time.Time      // Last failure timestamp
	providerName     string        // Provider name for logging
}

// NewCircuitBreaker creates a new CircuitBreaker instance
func NewCircuitBreaker(providerName string, failureThreshold int, cooldownPeriod time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		cooldownPeriod:   cooldownPeriod,
		state:            StateClosed,
		failures:         0,
		lastFailure:      time.Time{},
		providerName:     providerName,
	}
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Allow checks if a request should be allowed through
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.cooldownPeriod {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateClosed
		cb.failures = 0
	case StateClosed:
		cb.failures = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.failureThreshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.lastFailure = time.Time{}
}

// FailureCount returns the current failure count
func (cb *CircuitBreaker) FailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}
