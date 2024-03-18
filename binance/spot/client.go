package spot

import (
	"github.com/adshao/go-binance/v2"
)

type SpotClient struct {
	client *binance.Client
}

func NewClient(apiKey, secretKey string, UseTestnet bool) *SpotClient {
	binance.UseTestnet = UseTestnet
	return &SpotClient{
		client: binance.NewClient(apiKey, secretKey),
	}
}

func (c *SpotClient) GetClient() *binance.Client {
	return c.client
}
