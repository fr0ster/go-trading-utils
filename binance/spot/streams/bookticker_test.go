package streams_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/streams"
	streams_interface "github.com/fr0ster/go-trading-utils/interfaces/streams"
	"github.com/stretchr/testify/assert"
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

func TestInterface(t *testing.T) {
	test := func(u streams_interface.Stream) chan bool {
		return u.GetStreamEvent()
	}
	bts := streams.NewBookTickerStream("SUSHIUSDT")
	assert.NotNil(t, test(bts))
}
