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
	})
}
