package logg

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// timestampFormat is RFC 3339 with milliseconds and the real UTC offset
// (e.g. 2024-05-01T10:00:00.000+02:00, or ...Z when the server runs in UTC).
const timestampFormat = "2006-01-02T15:04:05.000Z07:00"

const dateLayout = "2006-01-02"

func Field(key string, value interface{}) LogField {
	return LogField{Key: key, Value: value}
}

// Converte valori come time.Duration in formati leggibili
func FormatValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Duration:
		return v.String()
	default:
		return v
	}
}

// GetCaller return filename + code line number of the caller.
func GetCaller() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	return file, line
}

func logFileName(logDir, prefix, date string) string {
	return logDir + "/" + prefix + "_" + date + ".log"
}

// GetLogFileName returns today's log file name for the given prefix
// (debug, info, warn, error).
func GetLogFileName(logDir, prefix string) string {
	return logFileName(logDir, prefix, time.Now().Format(dateLayout))
}

// openLogFiles opens the four level files for date. On error, the files
// already opened are closed.
func openLogFiles(logDir, date string) ([4]*os.File, error) {
	var files [4]*os.File
	for i, name := range levelNames {
		f, err := os.OpenFile(logFileName(logDir, name, date), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			for _, opened := range files[:i] {
				_ = opened.Close()
			}
			return [4]*os.File{}, err
		}
		files[i] = f
	}
	return files, nil
}

// useFiles points the per-level loggers to files.
func (l *Logger) useFiles(files [4]*os.File, date string) {
	for i, f := range files {
		var console io.Writer = os.Stdout
		if i == ERROR {
			console = os.Stderr
		}
		l.loggers[i] = log.New(io.MultiWriter(f, console), "", 0)
	}
	l.files = files
	l.curDate = date
}

func closeFiles(files [4]*os.File) {
	for i, f := range files {
		if f == nil {
			continue
		}
		if err := f.Close(); err != nil {
			log.Printf("Error while closing %s log file: %v", levelNames[i], err)
		}
	}
}

func NewLogger(level int, logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	today := time.Now().Format(dateLayout)
	files, err := openLogFiles(logDir, today)
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		level:      level,
		logDir:     logDir,
		logChannel: make(chan logEntry, 1000),
	}
	logger.useFiles(files, today)

	logger.wg.Add(1)
	go logger.startAsyncWriter()

	return logger, nil
}

func (l *Logger) startAsyncWriter() {
	defer l.wg.Done()
	for entry := range l.logChannel {
		l.writeLog(entry)
	}
}

// rotate switches to the files of date. Runs only in the writer goroutine.
// If the new files cannot be opened, it keeps writing to the current ones.
func (l *Logger) rotate(date string) {
	files, err := openLogFiles(l.logDir, date)
	if err != nil {
		log.Println("Error while rotating log files:", err)
		l.curDate = date // don't retry on every message; retry on the next date change
		return
	}
	old := l.files
	l.useFiles(files, date)
	closeFiles(old)
}

func (l *Logger) writeLog(entry logEntry) {
	if date := entry.ts.Format(dateLayout); date != l.curDate {
		l.rotate(date)
	}

	data := map[string]interface{}{
		"timestamp": entry.ts.Format(timestampFormat),
		"message":   entry.message,
	}
	for _, field := range entry.fields {
		if field.Value == nil {
			continue
		}
		if s, ok := field.Value.(string); ok && s == "" {
			continue
		}
		data[field.Key] = FormatValue(field.Value)
	}
	logJSON, err := json.Marshal(data)
	if err != nil {
		logJSON, _ = json.Marshal(map[string]interface{}{
			"timestamp":        entry.ts.Format(timestampFormat),
			"error":            "failed to marshal log message",
			"original_message": entry.message,
		})
	}
	if entry.level >= DEBUG && entry.level <= ERROR {
		l.loggers[entry.level].Println(string(logJSON))
	}
}

func (l *Logger) log(level int, message string, fields ...LogField) {
	if level >= l.level {
		entry := logEntry{level: level, ts: time.Now(), message: message, fields: fields}
		l.mu.RLock()
		defer l.mu.RUnlock()
		if l.closed {
			// Logger già chiuso: scarta il messaggio invece di andare in panic.
			return
		}
		select {
		case l.logChannel <- entry:
		default:
			// canale pieno, scarta silenziosamente
		}
	}
}

func (l *Logger) Debug(msg string, fields ...LogField) { l.log(DEBUG, msg, fields...) }
func (l *Logger) Info(msg string, fields ...LogField)  { l.log(INFO, msg, fields...) }
func (l *Logger) Warn(msg string, fields ...LogField)  { l.log(WARN, msg, fields...) }
func (l *Logger) Error(msg string, fields ...LogField) { l.log(ERROR, msg, fields...) }

// Close flushes the queued messages and closes the log files.
// It is safe to call more than once; messages logged after Close are dropped.
func (l *Logger) Close() {
	l.closeOnce.Do(func() {
		l.mu.Lock()
		l.closed = true
		close(l.logChannel)
		l.mu.Unlock()
		l.wg.Wait()
		closeFiles(l.files)
	})
}
