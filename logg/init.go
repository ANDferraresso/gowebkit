// Package logg provides helper functions logging.
package logg

import (
	"io"
	"log"
	"sync"
)

// Define log levels.
const (
	DEBUG = iota // 0
	INFO
	WARN
	ERROR
)

// Logger represents a log object.
type Logger struct {
	level      int
	debugLog   *log.Logger
	infoLog    *log.Logger
	warnLog    *log.Logger
	errorLog   *log.Logger
	debugFile  io.WriteCloser
	infoFile   io.WriteCloser
	warnFile   io.WriteCloser
	errorFile  io.WriteCloser
	logChannel chan logEntry
	wg         sync.WaitGroup
	closeOnce  sync.Once
}

// LogEntry represents a log row.
type logEntry struct {
	level   int
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
	"logg" // Assicurati che il modulo si chiami "logg" oppure usa il corretto path d'import
)

func main() {
	// Crea un logger con livello DEBUG e directory dei log
	logger, err := logg.NewLogger(logg.DEBUG, "./logs")
	if err != nil {
		panic(err)
	}
	defer logger.Close() // Assicurati di chiudere alla fine per svuotare il buffer

	start := time.Now()

	// Esempio di logging
	logger.Debug("Avvio del processo", logg.Field("modulo", "main"))
	logger.Info("Utente autenticato", logg.Field("user", "marco"))
	logger.Warn("Limite vicino", logg.Field("limite", 90), logg.Field("max", 100))
	logger.Error("Errore durante la lettura", logg.Field("file", "/tmp/data.json"), logg.Field("retry", true))

	// Logging con time.Duration
	elapsed := time.Since(start)
	logger.Info("Processo completato", logg.Field("durata", elapsed))

	// Attendi per assicurarti che i log vengano scritti prima di uscire (utile solo per demo/test)
	time.Sleep(500 * time.Millisecond)
}

*/
