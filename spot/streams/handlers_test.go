package streams_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/streams"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetFilledOrderHandler(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, secretKey)
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error starting user stream: %v", err)
	}
	_, _, err = streams.StartUserDataStream(listenKey, utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	executeOrderChan := streams.GetFilledOrdersGuard()

	assert.NotNil(t, executeOrderChan, "executeOrderChan should not be nil")
}

func TestGetBalanceTreeUpdateHandler(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, secretKey)
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error starting user stream: %v", err)
	}
	_, _, err = streams.StartUserDataStream(listenKey, utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	executeOrderChan := streams.GetBalancesUpdateGuard()

	assert.NotNil(t, executeOrderChan, "executeOrderChan should not be nil")

}

func TestGetBookTickersUpdateHandler(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, secretKey)
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error starting user stream: %v", err)
	}
	_, _, err = streams.StartBookTickerStream(listenKey, utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	bookTickerBoolChan := streams.GetBalancesUpdateGuard()

	assert.NotNil(t, bookTickerBoolChan, "Book Ticker channel should not be nil")
}

func TestGetDepthsUpdaterHandler(t *testing.T) {
	_, _, err := streams.StartDepthStream("BTCUSDT", utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	depthsChan := streams.GetDepthsUpdateGuard()
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	assert.NotNil(t, depthsChan, "Depths channel should not be nil")
}
