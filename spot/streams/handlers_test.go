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
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error starting user stream: %v", err)
	}
	_, _, err = streams.StartUserDataStream(listenKey, utils.HandleErr)
	if err != nil {
		log.Fatalf("Error starting user data stream: %v", err)
	}
	executeOrderChan := streams.GetFilledOrderHandler()

	assert.NotNil(t, executeOrderChan, "executeOrderChan should not be nil")
}
