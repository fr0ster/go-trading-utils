package streams_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/streams"
)

func TestNewKlineStream(t *testing.T) {
	stream := streams.NewDepthStream("BTCUSDT", false)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestKlineStream_Start(t *testing.T) {
	stream := streams.NewDepthStream("BTCUSDT", false)
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
