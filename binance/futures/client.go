package futures

import "github.com/adshao/go-binance/v2/futures"

type SpotClient struct {
	client *futures.Client
}

func NewClient(apiKey, secretKey string, UseTestnet bool) *SpotClient {
	futures.UseTestnet = UseTestnet
	return &SpotClient{
		client: futures.NewClient(apiKey, secretKey),
	}
}

func (c *SpotClient) GetClient() *futures.Client {
	return c.client
}
