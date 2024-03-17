package spot

import "github.com/adshao/go-binance/v2"

type SpotClient struct {
	client *binance.Client
}

func NewSpotClient(apiKey, secretKey string) *SpotClient {
	return &SpotClient{
		client: binance.NewClient(apiKey, secretKey),
	}
}

func (c *SpotClient) GetClient() *binance.Client {
	return c.client
}
