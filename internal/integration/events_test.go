package integration

import (
	"testing"
	"time"
)

func TestEventDispatcherMatchingHooks(t *testing.T) {
	hooks := []HookConfig{
		{Event: EventPromptSubmit, Command: "echo", Args: []string{"submit"}},
		{Event: EventRulesSelected, Command: "echo", Args: []string{"selected"}},
		{Event: EventPromptSubmit, Command: "echo", Args: []string{"submit2"}},
	}

	dispatcher := NewEventDispatcher(hooks)

	matched := dispatcher.MatchingHooks(EventPromptSubmit)
	if len(matched) != 2 {
		t.Errorf("expected 2 matching hooks, got %d", len(matched))
	}

	matched = dispatcher.MatchingHooks(EventRulesSelected)
	if len(matched) != 1 {
		t.Errorf("expected 1 matching hook, got %d", len(matched))
	}

	matched = dispatcher.MatchingHooks(EventSessionStart)
	if len(matched) != 0 {
		t.Errorf("expected 0 matching hooks, got %d", len(matched))
	}
}

func TestEventDispatcherHookCount(t *testing.T) {
	dispatcher := NewEventDispatcher([]HookConfig{
		{Event: EventPromptSubmit, Command: "echo"},
		{Event: EventRulesSelected, Command: "echo"},
	})

	if dispatcher.HookCount() != 2 {
		t.Errorf("HookCount() = %d, want 2", dispatcher.HookCount())
	}
}

func TestEventDispatcherDispatch(t *testing.T) {
	// Dispatch with "true" command (always succeeds, fire-and-forget)
	hooks := []HookConfig{
		{Event: EventPromptSubmit, Command: "true", Timeout: 2 * time.Second},
	}

	dispatcher := NewEventDispatcher(hooks)
	payload := EventPayload{
		Project: "/test/project",
		Data:    map[string]string{"prompt": "test"},
	}

	// Should not panic or block
	dispatcher.Dispatch(EventPromptSubmit, payload)

	// Give goroutine time to execute
	time.Sleep(100 * time.Millisecond)
}

func TestFormatPayload(t *testing.T) {
	payload := EventPayload{
		Event:   EventRulesSelected,
		Project: "/test",
		Data:    []string{"security.md", "react.md"},
	}

	result, err := FormatPayload(payload)
	if err != nil {
		t.Fatalf("FormatPayload error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestEventTypes(t *testing.T) {
	// Verify event type constants
	if EventPromptSubmit != "onPromptSubmit" {
		t.Error("EventPromptSubmit constant wrong")
	}
	if EventRulesSelected != "onRulesSelected" {
		t.Error("EventRulesSelected constant wrong")
	}
	if EventSessionStart != "onSessionStart" {
		t.Error("EventSessionStart constant wrong")
	}
	if EventToolDiscovered != "onToolDiscovered" {
		t.Error("EventToolDiscovered constant wrong")
	}
}
