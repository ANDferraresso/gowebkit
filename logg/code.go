package logg

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

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

// getCaller restituisce il nome del file e la linea chiamante
func GetCaller() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	return file, line
}

func getLogFileName(logDir, prefix string) string {
	today := time.Now().Format("2006-01-02")
	return logDir + "/" + prefix + "_" + today + ".log"
}

func GetLogFileName(logDir, prefix string) string {
	return getLogFileName(logDir, prefix)
}

func NewLogger(level int, logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	df, err := os.OpenFile(getLogFileName(logDir, "debug"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	inf, err := os.OpenFile(getLogFileName(logDir, "info"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	wf, err := os.OpenFile(getLogFileName(logDir, "warn"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	ef, err := os.OpenFile(getLogFileName(logDir, "error"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		level:      level,
		debugLog:   log.New(io.MultiWriter(df, os.Stdout), "", 0),
		infoLog:    log.New(io.MultiWriter(inf, os.Stdout), "", 0),
		warnLog:    log.New(io.MultiWriter(wf, os.Stdout), "", 0),
		errorLog:   log.New(io.MultiWriter(ef, os.Stderr), "", 0),
		debugFile:  df,
		infoFile:   inf,
		warnFile:   wf,
		errorFile:  ef,
		logChannel: make(chan logEntry, 1000),
	}

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

func (l *Logger) writeLog(entry logEntry) {
	data := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02T15:04:05.000Z"),
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
			"error":            "failed to marshal log message",
			"original_message": entry.message,
		})
	}
	switch entry.level {
	case DEBUG:
		l.debugLog.Println(string(logJSON))
	case INFO:
		l.infoLog.Println(string(logJSON))
	case WARN:
		l.warnLog.Println(string(logJSON))
	case ERROR:
		l.errorLog.Println(string(logJSON))
	}
}

func (l *Logger) log(level int, message string, fields ...LogField) {
	if level >= l.level {
		entry := logEntry{level: level, message: message, fields: fields}
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

func (l *Logger) Close() {
	l.closeOnce.Do(func() {
		close(l.logChannel)
		l.wg.Wait()
		err := l.debugFile.Close()
		if err != nil {
			fmt.Println("Error while l.debugFile.Close()", err)
		}
		err = l.infoFile.Close()
		if err != nil {
			fmt.Println("Error while l.infoFile.Close()", err)
		}
		err = l.warnFile.Close()
		if err != nil {
			fmt.Println("Error while l.warnFile.Close()", err)
		}
		err = l.errorFile.Close()
		if err != nil {
			fmt.Println("Error while l.errorFile.Close()", err)
		}
	})
}
