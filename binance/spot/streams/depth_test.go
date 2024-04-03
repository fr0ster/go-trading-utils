package streams_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/streams"
)

func TestNewDepthStream(t *testing.T) {
	stream := streams.NewDepthStream("BTCUSDT", true)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestDepthStream_Start(t *testing.T) {
	stream := streams.NewDepthStream("BTCUSDT", true)
	doneC, stopC, err := stream.Start()
	if err != nil {
		t.Error(err)
	}
	if doneC == nil {
		t.Error("Expected not nil")
	}
	if stopC == nil {
		t.Error("Expected not nil")
	}
}
