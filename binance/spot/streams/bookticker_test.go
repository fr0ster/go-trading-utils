package streams_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/streams"
)

func TestNewBookTickerStream(t *testing.T) {
	symbol := "SUSHIUSDT"
	stream := streams.NewBookTickerStream(symbol)
	if stream == nil {
		t.Error("Expected stream to be created")
	}
}

func TestBookTickerStream_Start(t *testing.T) {
	symbol := "SUSHIUSDT"
	stream := streams.NewBookTickerStream(symbol)
	doneC, stopC, err := stream.Start()
	if err != nil {
		t.Error(err)
	}
	if doneC == nil {
		t.Error("Expected doneC to be created")
	}
	if stopC == nil {
		t.Error("Expected stopC to be created")
	}
}
