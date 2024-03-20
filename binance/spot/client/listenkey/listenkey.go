package listenkey

import (
	"context"

	"github.com/adshao/go-binance/v2"
)

type (
	SpotClient struct {
		client *binance.Client
	}
)

func New(api_key, secret_key string, UseTestnet bool) *SpotClient {
	binance.UseTestnet = UseTestnet
	return &SpotClient{binance.NewClient(api_key, secret_key)}
}

func (c *SpotClient) GetListenKey() (string, error) {
	return c.client.NewStartUserStreamService().Do(context.Background())
}
