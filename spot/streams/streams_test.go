package streams_test

import (
	"context"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/streams"
	"github.com/fr0ster/go-binance-utils/utils"
)

func TestStartUserDataStream(t *testing.T) {
	t.Run("StartUserDataStream", func(t *testing.T) {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		listenKey, err := client.NewStartUserStreamService().Do(context.Background())
		if err != nil {
			t.Fatalf("Error starting user stream: %v", err)
		}
		doneC, stopC, err := streams.StartUserDataStream(listenKey, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}

		channel, res := streams.GetUserDataChannel()
		if !res || channel == nil {
			t.Error("Failed to get user data channel")
		}
	})

}

func TestStartDepthStream(t *testing.T) {
	t.Run("StartDepthStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		doneC, stopC, err := streams.StartDepthStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}

		channel, res := streams.GetDepthChannel()
		if !res || channel == nil {
			t.Error("Failed to get depth channel")
		}
	})
}

func TestStartKlineStream(t *testing.T) {
	t.Run("StartKlineStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		doneC, stopC, err := streams.StartKlineStream(symbol, "1m", utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}

		channel, res := streams.GetKlineChannel()
		if !res || channel == nil {
			t.Error("Failed to get kline channel")
		}
	})
}

func TestStartTradeStream(t *testing.T) {
	t.Run("StartTradeStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		doneC, stopC, err := streams.StartTradeStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}

		channel, res := streams.GetTradeChannel()
		if !res || channel == nil {
			t.Error("Failed to get trade channel")
		}
	})
}

func TestStartAggTradeStream(t *testing.T) {
	t.Run("StartAggTradeStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		doneC, stopC, err := streams.StartAggTradeStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}

		channel, res := streams.GetAggTradeChannel()
		if !res || channel == nil {
			t.Error("Failed to get aggregated trade channel")
		}
	})
}

func TestStartBookTickerStream(t *testing.T) {
	t.Run("StartBookTickerStream", func(t *testing.T) {
		symbol := "BTCUSDT"
		doneC, stopC, err := streams.StartBookTickerStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if doneC == nil {
			t.Error("doneC is nil")
		}

		if stopC == nil {
			t.Error("stopC is nil")
		}

		channel, res := streams.GetBookTickerChannel()
		if !res || channel == nil {
			t.Error("Failed to get book ticker channel")
		}
	})
}

func TestGetUserDataChannel(t *testing.T) {
	t.Run("GetUserDataChannel", func(t *testing.T) {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		listenKey, err := client.NewStartUserStreamService().Do(context.Background())
		if err != nil {
			t.Fatalf("Error starting user stream: %v", err)
		}
		_, _, err = streams.StartUserDataStream(listenKey, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		channel, ok := streams.GetUserDataChannel()

		if !ok {
			t.Error("Failed to get user data channel")
		}

		if channel == nil {
			t.Error("User data channel is nil")
		}
	})
}

func TestGetDepthChannel(t *testing.T) {
	t.Run("GetDepthChannel", func(t *testing.T) {
		symbol := "BTCUSDT"
		_, _, err := streams.StartDepthStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		channel, ok := streams.GetDepthChannel()

		if !ok {
			t.Error("Failed to get depth channel")
		}

		if channel == nil {
			t.Error("Depth channel is nil")
		}
	})
}

func TestGetKlineChannel(t *testing.T) {
	t.Run("GetKlineChannel", func(t *testing.T) {
		symbol := "BTCUSDT"
		_, _, err := streams.StartKlineStream(symbol, "1m", utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		channel, ok := streams.GetKlineChannel()

		if !ok {
			t.Error("Failed to get kline channel")
		}

		if channel == nil {
			t.Error("Kline channel is nil")
		}
	})
}

func TestGetTradeChannel(t *testing.T) {
	t.Run("GetTradeChannel", func(t *testing.T) {
		symbol := "BTCUSDT"
		_, _, err := streams.StartTradeStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		channel, ok := streams.GetTradeChannel()

		if !ok {
			t.Error("Failed to get trade channel")
		}

		if channel == nil {
			t.Error("Trade channel is nil")
		}
	})
}

func TestGetAggTradeChannel(t *testing.T) {
	t.Run("GetAggTradeChannel", func(t *testing.T) {
		symbol := "BTCUSDT"
		_, _, err := streams.StartAggTradeStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		channel, ok := streams.GetAggTradeChannel()

		if !ok {
			t.Error("Failed to get aggregated trade channel")
		}

		if channel == nil {
			t.Error("Aggregated trade channel is nil")
		}
	})
}

func TestGetBookTickerChannel(t *testing.T) {
	t.Run("GetBookTickerChannel", func(t *testing.T) {
		symbol := "BTCUSDT"
		_, _, err := streams.StartBookTickerStream(symbol, utils.HandleErr)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		channel, ok := streams.GetBookTickerChannel()

		if !ok {
			t.Error("Failed to get book ticker channel")
		}

		if channel == nil {
			t.Error("Book ticker channel is nil")
		}
	})
}
