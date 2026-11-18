package logg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Entries are written to the file of the day they were logged, and the
// timestamp carries the real UTC offset.
func TestRotationAndTimestamp(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(ERROR+1, dir) // nothing logged through the public API
	if err != nil {
		t.Fatal(err)
	}

	zone := time.FixedZone("UTC+2", 2*60*60)
	tomorrow := time.Now().In(zone).Add(24 * time.Hour)
	l.mu.RLock()
	l.logChannel <- logEntry{level: INFO, ts: tomorrow, message: "next day"}
	l.mu.RUnlock()
	l.Close()

	name := logFileName(dir, "info", tomorrow.Format(dateLayout))
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("rotated file not created: %v", err)
	}
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["message"] != "next day" {
		t.Errorf("message = %v", entry["message"])
	}
	ts, _ := entry["timestamp"].(string)
	if !strings.HasSuffix(ts, "+02:00") {
		t.Errorf("timestamp %q does not carry the +02:00 offset", ts)
	}
	parsed, err := time.Parse(timestampFormat, ts)
	if err != nil || !parsed.Equal(tomorrow.Truncate(time.Millisecond)) {
		t.Errorf("timestamp %q parses to %v (err %v), want %v", ts, parsed, err, tomorrow)
	}

	// Today's files exist too (opened at startup) and were closed.
	if _, err := os.Stat(filepath.Join(dir, "info_"+time.Now().Format(dateLayout)+".log")); err != nil {
		t.Errorf("today's file missing: %v", err)
	}
}

func TestNewLoggerFailsCleanly(t *testing.T) {
	dir := t.TempDir()
	// Make the "warn" file impossible to open: a directory with that name.
	today := time.Now().Format(dateLayout)
	if err := os.Mkdir(logFileName(dir, "warn", today), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := NewLogger(DEBUG, dir); err == nil {
		t.Fatal("NewLogger succeeded with an unopenable file")
	}
}
