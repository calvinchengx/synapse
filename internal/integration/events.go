package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// EventType identifies a lifecycle event.
type EventType string

const (
	EventPromptSubmit   EventType = "onPromptSubmit"
	EventRulesSelected  EventType = "onRulesSelected"
	EventSessionStart   EventType = "onSessionStart"
	EventToolDiscovered EventType = "onToolDiscovered"
)

// EventPayload is the JSON payload sent to hook commands.
type EventPayload struct {
	Event   EventType `json:"event"`
	Data    any       `json:"data,omitempty"`
	Project string    `json:"project,omitempty"`
}

// HookConfig describes a registered event hook.
type HookConfig struct {
	Event   EventType
	Command string
	Args    []string
	Timeout time.Duration
}

// EventDispatcher manages fire-and-forget event dispatch to registered hooks.
type EventDispatcher struct {
	hooks []HookConfig
}

// NewEventDispatcher creates an EventDispatcher with the given hooks.
func NewEventDispatcher(hooks []HookConfig) *EventDispatcher {
	return &EventDispatcher{hooks: hooks}
}

// Dispatch fires an event to all matching hooks. Never blocks the caller.
// Errors are logged but never returned.
func (d *EventDispatcher) Dispatch(event EventType, payload EventPayload) {
	payload.Event = event
	for _, hook := range d.hooks {
		if hook.Event != event {
			continue
		}
		go d.runHook(hook, payload)
	}
}

// runHook executes a single hook command with the payload on stdin.
func (d *EventDispatcher) runHook(hook HookConfig, payload EventPayload) {
	timeout := hook.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, hook.Command, hook.Args...)

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	cmd.Stdin = &byteReader{data: data}
	_ = cmd.Run() // fire-and-forget: ignore errors
}

// MatchingHooks returns hooks registered for the given event.
func (d *EventDispatcher) MatchingHooks(event EventType) []HookConfig {
	var matched []HookConfig
	for _, h := range d.hooks {
		if h.Event == event {
			matched = append(matched, h)
		}
	}
	return matched
}

// HookCount returns the total number of registered hooks.
func (d *EventDispatcher) HookCount() int {
	return len(d.hooks)
}

// FormatPayload serializes an EventPayload to JSON.
func FormatPayload(payload EventPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshaling payload: %w", err)
	}
	return string(data), nil
}
