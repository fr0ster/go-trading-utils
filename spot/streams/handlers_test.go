package streams_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/fr0ster/go-binance-utils/spot/streams"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetFilledOrderHandler(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
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
	// binance.UseTestnet = true
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
	_, _, err := streams.StartBookTickerStream("BTCUSDT", utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	bookTickerBoolChan := streams.GetBalancesUpdateGuard()

	assert.NotNil(t, bookTickerBoolChan, "Book Ticker channel should not be nil")
}

func getTestDepths() *markets.DepthBTree {
	testDepthTree := markets.DepthNew(3)
	records := []markets.DepthItem{
		{Price: 1.92, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 150.2},
		{Price: 1.93, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 155.4}, // local maxima
		{Price: 1.94, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 150.0},
		{Price: 1.941, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 130.4},
		{Price: 1.947, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 172.1},
		{Price: 1.948, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 187.4},
		{Price: 1.949, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 236.1}, // local maxima
		{Price: 1.95, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 189.8},
		{Price: 1.951, AskLastUpdateID: 2369068, AskQuantity: 217.9, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.952, AskLastUpdateID: 2369068, AskQuantity: 179.4, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.953, AskLastUpdateID: 2369068, AskQuantity: 180.9, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.954, AskLastUpdateID: 2369068, AskQuantity: 148.5, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.955, AskLastUpdateID: 2369068, AskQuantity: 120.0, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.956, AskLastUpdateID: 2369068, AskQuantity: 110.0, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.957, AskLastUpdateID: 2369068, AskQuantity: 140.0, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.958, AskLastUpdateID: 2369068, AskQuantity: 90.0, BidLastUpdateID: 0, BidQuantity: 0},
	}
	for _, record := range records {
		testDepthTree.ReplaceOrInsert(&record)
	}

	return testDepthTree
}

func TestGetDepthsUpdaterHandler(t *testing.T) {
	_, _, err := streams.StartDepthStream("BTCUSDT", utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	depthsChan := streams.GetDepthsUpdateGuard(getTestDepths())
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	assert.NotNil(t, depthsChan, "Depths channel should not be nil")
}
