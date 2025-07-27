package logg

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func readLastLogLine(filename string) (map[string]interface{}, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		return nil, nil
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(lines[len(lines)-1]), &result)
	return result, err
}

func TestAsyncLoggerWritesToAllLevels(t *testing.T) {
	tmpDir := t.TempDir()

	logger, err := NewLogger(DEBUG, tmpDir)
	if err != nil {
		t.Fatalf("Error while creating logger: %v", err)
	}
	defer logger.Close()

	tests := []struct {
		level    int
		message  string
		fileType string
	}{
		{DEBUG, "Log debug", "debug"},
		{INFO, "Log info", "info"},
		{WARN, "Log warn", "warn"},
		{ERROR, "Log error", "error"},
	}

	for _, tt := range tests {
		switch tt.level {
		case DEBUG:
			logger.Debug(tt.message, Field("test", true))
		case INFO:
			logger.Info(tt.message, Field("test", true))
		case WARN:
			logger.Warn(tt.message, Field("test", true))
		case ERROR:
			logger.Error(tt.message, Field("test", true))
		}
	}

	logger.Close() // assicura che i log siano scritti

	for _, tt := range tests {
		logFile := GetLogFileName(tmpDir, tt.fileType)
		entry, err := readLastLogLine(logFile)
		if err != nil {
			t.Errorf("Errore while reading %s: %v", tt.fileType, err)
			continue
		}
		if entry["message"] != tt.message {
			t.Errorf("Log message %s wrong: %v", tt.fileType, entry["message"])
		}
	}
}

func TestLoggerWithFullChannel(t *testing.T) {
	tmpDir := t.TempDir()

	logger, err := NewLogger(INFO, tmpDir)
	if err != nil {
		t.Fatalf("Error while creating logger: %v", err)
	}
	//defer logger.Close()

	for i := 0; i < 2000; i++ { // supera la capacità del canale
		logger.Info("Message", Field("i", i))
	}
	logger.Close()
	// Se il test arriva qui senza blocchi o panico, il logger è robusto
}
