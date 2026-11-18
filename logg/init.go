// Package logg provides helper functions logging.
package logg

import (
	"log"
	"os"
	"sync"
	"time"
)

// Define log levels.
const (
	DEBUG = iota // 0
	INFO
	WARN
	ERROR
)

// levelNames are the file prefixes, indexed by level.
var levelNames = [4]string{"debug", "info", "warn", "error"}

// Logger represents a log object.
//
// Each level is written as JSON lines to its own file, <logDir>/<level>_<YYYY-MM-DD>.log,
// and to stdout (stderr for ERROR). Files rotate automatically when the date changes.
type Logger struct {
	level      int
	logDir     string
	curDate    string // date of the files currently open (owned by the writer goroutine)
	loggers    [4]*log.Logger
	files      [4]*os.File
	logChannel chan logEntry
	wg         sync.WaitGroup
	closeOnce  sync.Once
	mu         sync.RWMutex // guards closed and sends on logChannel
	closed     bool
}

// LogEntry represents a log row.
type logEntry struct {
	level   int
	ts      time.Time
	message string
	fields  []LogField
}

// LogField represents a key-value pair in a log row.
type LogField struct {
	Key   string
	Value interface{}
}

/*

Example:

package main

import (
	"time"

	"github.com/ANDferraresso/gowebkit/logg"
)

func main() {
	// Crea un logger con livello DEBUG e directory dei log
	logger, err := logg.NewLogger(logg.DEBUG, "./logs")
	if err != nil {
		panic(err)
	}
	defer logger.Close() // Chiude e svuota il buffer: i messaggi in coda vengono scritti

	start := time.Now()

	// Esempio di logging
	logger.Debug("Avvio del processo", logg.Field("modulo", "main"))
	logger.Info("Utente autenticato", logg.Field("user", "marco"))
	logger.Warn("Limite vicino", logg.Field("limite", 90), logg.Field("max", 100))
	logger.Error("Errore durante la lettura", logg.Field("file", "/tmp/data.json"), logg.Field("retry", true))

	// Logging con time.Duration
	elapsed := time.Since(start)
	logger.Info("Processo completato", logg.Field("durata", elapsed))
}

*/
