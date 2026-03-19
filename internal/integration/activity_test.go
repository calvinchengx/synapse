package integration

import (
	"path/filepath"
	"testing"
	"time"
)

func TestActivityLogWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity.jsonl")
	log := NewActivityLog(path)

	// Write entries
	entries := []ActivityEntry{
		{
			Timestamp: time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC),
			Event:     "onRulesSelected",
			Rules:     []string{"security.md", "react.md"},
			Method:    "semantic",
			Project:   "/path/to/project",
		},
		{
			Timestamp: time.Date(2026, 3, 29, 10, 5, 0, 0, time.UTC),
			Event:     "onRulesSelected",
			Rules:     []string{"testing.md"},
			Method:    "keyword",
			Project:   "/path/to/project",
		},
	}

	for _, e := range entries {
		if err := log.Log(e); err != nil {
			t.Fatalf("Log error: %v", err)
		}
	}

	// Read back
	read, err := log.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	if len(read) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(read))
	}
	if read[0].Event != "onRulesSelected" {
		t.Errorf("event = %q", read[0].Event)
	}
	if len(read[0].Rules) != 2 {
		t.Errorf("expected 2 rules in first entry, got %d", len(read[0].Rules))
	}
	if read[1].Method != "keyword" {
		t.Errorf("method = %q", read[1].Method)
	}
}

func TestActivityLogReadEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")
	log := NewActivityLog(path)

	entries, err := log.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestActivityLogAutoTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity.jsonl")
	log := NewActivityLog(path)

	// Log without timestamp
	err := log.Log(ActivityEntry{
		Event: "onSessionStart",
	})
	if err != nil {
		t.Fatalf("Log error: %v", err)
	}

	entries, err := log.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Timestamp.IsZero() {
		t.Error("expected auto-populated timestamp")
	}
}
