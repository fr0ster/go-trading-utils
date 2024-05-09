package streams_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/futures/deprecated/streams"
)

func TestNewPartialDepthStream(t *testing.T) {
	stream := streams.NewPartialDepthStream("BTCUSDT", 5, 1)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestNewPartialDepthStream_Start(t *testing.T) {
	stream := streams.NewPartialDepthStream("BTCUSDT", 5, 1)
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

func TestNewDiffDepthStream(t *testing.T) {
	stream := streams.NewDiffDepthStream("BTCUSDT", 1)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestNewDiffDepthStream_Start(t *testing.T) {
	stream := streams.NewDiffDepthStream("BTCUSDT", 1)
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

func TestNewPartialDepthStreamWithRate(t *testing.T) {
	stream := streams.NewPartialDepthStreamWithRate("BTCUSDT", 5, streams.Rate100Ms, 1)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestNewPartialDepthStreamWithRate_Start(t *testing.T) {
	stream := streams.NewPartialDepthStreamWithRate("BTCUSDT", 5, streams.Rate100Ms, 1)
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

func TestNewCombinedDepthStream(t *testing.T) {
	symbols := make(map[string]string) // Initialize the map
	symbols["BTCUSDT"] = "BTCUSDT"
	stream := streams.NewCombinedDepthStream(symbols, 1)
	if stream == nil {
		t.Error("Expected not nil")
	}
}

func TestNewCombinedDepthStream_Start(t *testing.T) {
	symbols := make(map[string]string) // Initialize the map
	symbols["BTCUSDT"] = "BTCUSDT"
	stream := streams.NewCombinedDepthStream(symbols, 1)
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
