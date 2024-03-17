package utils_test

import (
	"os"
	"testing"
	"time"
)

func TestHandleShutdown(t *testing.T) {
	stop := make(chan os.Signal)
	delay := time.Second

	go func() {
		// Simulate receiving a signal after the specified delay
		time.Sleep(delay)
		stop <- os.Interrupt
	}()

	select {
	case <-stop:
		// Expected behavior
	case <-time.After(delay + time.Second):
		t.Error("HandleShutdown did not receive the expected signal")
	}
}
