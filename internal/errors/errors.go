package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ExternalError represents a failure when interacting with an external resource.
type ExternalError struct {
	Source  string // "rtk", "agentsview", "github", "llm"
	Op      string // "query", "connect", "fetch", "open"
	Err     error
	Timeout bool
}

func (e *ExternalError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("%s %s: timeout: %v", e.Source, e.Op, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Source, e.Op, e.Err)
}

func (e *ExternalError) Unwrap() error {
	return e.Err
}

// NewExternalError creates a new ExternalError.
func NewExternalError(source, op string, err error) *ExternalError {
	return &ExternalError{Source: source, Op: op, Err: err}
}

// NewTimeoutError creates a new ExternalError with Timeout set to true.
func NewTimeoutError(source, op string, err error) *ExternalError {
	return &ExternalError{Source: source, Op: op, Err: err, Timeout: true}
}

// AsExternalError unwraps err into an *ExternalError if possible.
func AsExternalError(err error, target **ExternalError) bool {
	return errors.As(err, target)
}

// IsTimeout reports whether the error is a timeout from an external resource.
func IsTimeout(err error) bool {
	var extErr *ExternalError
	if errors.As(err, &extErr) {
		return extErr.Timeout
	}
	return false
}

// IsLocked reports whether the error indicates a locked SQLite database.
func IsLocked(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "SQLITE_BUSY")
}

// State represents the state of a circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation, requests pass through
	StateOpen                  // failures exceeded threshold, requests blocked
	StateHalfOpen              // testing if service recovered
)

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

// CircuitBreaker prevents repeated calls to a failing external service.
type CircuitBreaker struct {
	mu          sync.Mutex
	failures    int
	threshold   int
	resetAfter  time.Duration
	lastFailure time.Time
	state       State
}

// NewCircuitBreaker creates a circuit breaker that opens after threshold
// consecutive failures and resets after resetAfter duration.
func NewCircuitBreaker(threshold int, resetAfter time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:  threshold,
		resetAfter: resetAfter,
		state:      StateClosed,
	}
}

// Allow reports whether a request should be allowed through.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.resetAfter {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return true
	}
}

// RecordSuccess records a successful call. Resets the circuit breaker to closed.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure records a failed call. Opens the circuit if threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = StateOpen
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen && time.Since(cb.lastFailure) >= cb.resetAfter {
		cb.state = StateHalfOpen
	}
	return cb.state
}

// Failures returns the current consecutive failure count.
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}
