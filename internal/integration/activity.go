package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ActivityLog writes append-only JSONL entries for rule selection events.
type ActivityLog struct {
	mu   sync.Mutex
	path string
}

// NewActivityLog creates an ActivityLog that writes to the given path.
func NewActivityLog(path string) *ActivityLog {
	return &ActivityLog{path: path}
}

// Log appends an activity entry to the log file.
func (al *ActivityLog) Log(entry ActivityEntry) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling activity entry: %w", err)
	}
	data = append(data, '\n')

	f, err := os.OpenFile(al.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening activity log: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

// Read reads all entries from the activity log.
func (al *ActivityLog) Read() ([]ActivityEntry, error) {
	data, err := os.ReadFile(al.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []ActivityEntry
	decoder := json.NewDecoder(
		&byteReader{data: data, pos: 0},
	)
	for decoder.More() {
		var entry ActivityEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// byteReader wraps a byte slice to implement io.Reader.
type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
