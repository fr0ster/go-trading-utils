package streams_test

import (
	"testing"

	"github.com/fr0ster/go-binance-utils/spot/streams"
)

func TestGetFilledOrderHandler(t *testing.T) {
	// Call the function under test
	wsHandler, _ := streams.GetFilledOrderHandler()

	// Verify that the returned handler function is not nil
	if wsHandler == nil {
		t.Error("Expected non-nil handler function, got nil")
	}
}
