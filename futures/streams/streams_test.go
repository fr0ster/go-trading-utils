package streams_test

import (
	"context"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-binance-utils/futures/streams"
	"github.com/fr0ster/go-binance-utils/utils"
)

func TestStartUserDataStream(t *testing.T) {
	t.Run("StartUserDataStream", func(t *testing.T) {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		// futures.UseTestnet = true
		client := futures.NewClient(api_key, secret_key)
		listenKey, err := client.NewStartUserStreamService().Do(context.Background())
		if err != nil {
			t.Fatalf("Error starting user stream: %v", err)
		}
		eventCh := make(chan *futures.WsUserDataEvent, 1)
		doneC, stopC, err := streams.StartUserDataStream(listenKey, eventCh, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}
	})

}

func TestStartDepthStream(t *testing.T) {
	t.Run("StartDepthStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		eventCh := make(chan *futures.WsDepthEvent, 5)
		doneC, stopC, err := streams.StartPartialDepthStream(symbol, 5, eventCh, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}
	})
}

func TestStartKlineStream(t *testing.T) {
	t.Run("StartKlineStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		eventCh := make(chan *futures.WsKlineEvent, 1)
		doneC, stopC, err := streams.StartKlineStream(symbol, "1m", eventCh, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}
	})
}

func TestStartCombinedAggTradeStream(t *testing.T) {
	t.Run("StartTradeStream", func(t *testing.T) {
		symbols := append([]string{"BTCUSDT"}, "ETHUSDT")
		eventCh := make(chan *futures.WsAggTradeEvent, 1)
		doneC, stopC, err := streams.StartCombinedAggTradeStream(symbols, eventCh, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}
	})
}

func TestStartAggTradeStream(t *testing.T) {
	t.Run("StartAggTradeStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		eventCh := make(chan *futures.WsAggTradeEvent, 1)
		doneC, stopC, err := streams.StartAggTradeStream(symbol, eventCh, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}
	})
}

func TestStartBookTickerStream(t *testing.T) {
	t.Run("StartBookTickerStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		eventCh := make(chan *futures.WsBookTickerEvent, 1)
		doneC, stopC, err := streams.StartBookTickerStream(symbol, eventCh, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}
	})
}
