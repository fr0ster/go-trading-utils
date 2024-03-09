package streams

import (
	"testing"
)

func TestGetFilledOrderHandler(t *testing.T) {
	// Call the function under test
	wsHandler, _ := GetFilledOrderHandler()

	// Verify that the returned handler function is not nil
	if wsHandler == nil {
		t.Error("Expected non-nil handler function, got nil")
	}
}
