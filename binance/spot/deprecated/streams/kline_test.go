package streams_test

import (
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/deprecated/streams"
)

func TestNewKlineStream(t *testing.T) {
	stream := streams.NewDepthStream("BTCUSDT", false, 1)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestKlineStream_Start(t *testing.T) {
	stream := streams.NewKlineStream("BTCUSDT", "1m", 1)
	doneC, stopC, err := stream.Start(func(event *binance.WsKlineEvent) {
		t.Log(event)
	})
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
