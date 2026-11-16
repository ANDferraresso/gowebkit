package logg

import (
	"sync"
	"testing"
)

// Logging after Close (or concurrently with it) must not panic.
func TestLogAfterCloseDoesNotPanic(t *testing.T) {
	l, err := NewLogger(ERROR+1, t.TempDir()) // level above ERROR: nothing is printed
	if err != nil {
		t.Fatal(err)
	}
	// Enable all levels so the calls below reach the channel send path;
	// nothing is logged before Close, so stdout stays quiet.
	l.level = DEBUG
	l.Close()
	l.Info("after close")
	l.Error("after close")
	l.Close() // idempotent
}

func TestLogConcurrentWithClose(t *testing.T) {
	l, err := NewLogger(ERROR+1, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				l.Debug("x")
			}
		}()
	}
	l.Close()
	wg.Wait()
}
