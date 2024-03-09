package utils

import (
	"context"
	"log"

	"github.com/adshao/go-binance/v2"
)

func CreateBinanceClient(apiKey, secretKey string) *binance.Client {
	binance.UseTestnet = true
	return binance.NewClient(apiKey, secretKey)
}

func GetListenKey(client *binance.Client) string {
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error starting user stream: %v", err)
	}
	return listenKey
}
