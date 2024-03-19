package listenkey

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
)

type (
	SpotClient struct {
		client *futures.Client
	}
)

func New(api_key, secret_key string, UseTestnet bool) *SpotClient {
	futures.UseTestnet = UseTestnet
	return &SpotClient{futures.NewClient(api_key, secret_key)}
}

func (c *SpotClient) GetListenKey() (string, error) {
	return c.client.NewStartUserStreamService().Do(context.Background())
}
