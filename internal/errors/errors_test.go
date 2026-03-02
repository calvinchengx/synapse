package errors

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestExternalError(t *testing.T) {
	inner := fmt.Errorf("connection refused")
	err := NewExternalError("agentsview", "connect", inner)

	if err.Error() != "agentsview connect: connection refused" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
	if err.Timeout {
		t.Error("expected Timeout to be false")
	}

	// Unwrap
	if !errors.Is(err, inner) {
		t.Error("expected Unwrap to return inner error")
	}
}

func TestTimeoutError(t *testing.T) {
	inner := fmt.Errorf("deadline exceeded")
	err := NewTimeoutError("llm", "fetch", inner)

	if err.Error() != "llm fetch: timeout: deadline exceeded" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
	if !err.Timeout {
		t.Error("expected Timeout to be true")
	}
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"plain error", fmt.Errorf("something"), false},
		{"external non-timeout", NewExternalError("rtk", "query", fmt.Errorf("fail")), false},
		{"external timeout", NewTimeoutError("llm", "fetch", fmt.Errorf("deadline")), true},
		{"wrapped timeout", fmt.Errorf("wrap: %w", NewTimeoutError("llm", "fetch", fmt.Errorf("deadline"))), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeout(tt.err); got != tt.expected {
				t.Errorf("IsTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsLocked(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"plain error", fmt.Errorf("something"), false},
		{"database is locked", fmt.Errorf("database is locked"), true},
		{"SQLITE_BUSY", fmt.Errorf("SQLITE_BUSY"), true},
		{"wrapped locked", fmt.Errorf("query failed: %w", fmt.Errorf("database is locked")), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocked(tt.err); got != tt.expected {
				t.Errorf("IsLocked() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %s, want %s", tt.state, got, tt.expected)
		}
	}
}

func TestCircuitBreakerClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, 60*time.Second)

	if !cb.Allow() {
		t.Error("expected Allow() = true when closed")
	}
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %s", cb.State())
	}
}

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Error("expected StateClosed after 2 failures (threshold=3)")
	}

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen after 3 failures, got %s", cb.State())
	}

	if cb.Allow() {
		t.Error("expected Allow() = false when open")
	}
}

func TestCircuitBreakerResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	if cb.Failures() != 0 {
		t.Errorf("expected 0 failures after success, got %d", cb.Failures())
	}
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after success, got %s", cb.State())
	}
}

func TestCircuitBreakerHalfOpenAfterReset(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %s", cb.State())
	}

	// Wait for reset duration
	time.Sleep(60 * time.Millisecond)

	// Should transition to half-open
	if cb.State() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen after reset duration, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected Allow() = true when half-open")
	}

	// Success in half-open → closed
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after success in half-open, got %s", cb.State())
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)

	// Open the circuit
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %s", cb.State())
	}

	// Wait for reset
	time.Sleep(60 * time.Millisecond)

	// Half-open: allow one request
	if !cb.Allow() {
		t.Fatal("expected Allow() = true when half-open")
	}

	// Failure in half-open → back to open
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen after failure in half-open, got %s", cb.State())
	}
}
