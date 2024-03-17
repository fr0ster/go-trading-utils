package spot

import "github.com/adshao/go-binance/v2/futures"

type SpotClient struct {
	client *futures.Client
}

func NewSpotClient(apiKey, secretKey string) *SpotClient {
	return &SpotClient{
		client: futures.NewClient(apiKey, secretKey),
	}
}

func (c *SpotClient) GetClient() *futures.Client {
	return c.client
}
