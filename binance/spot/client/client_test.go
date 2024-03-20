package client_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/client"
	client_interface "github.com/fr0ster/go-trading-utils/interfaces/client"
)

func TestInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	spot := client.New(api_key, secret_key, "BTCUSDT", 5, UseTestnet)
	err := func(client client_interface.Client) error {
		res, err := client.NewDepthServiceDo()
		fmt.Println(res, err)
		return err
	}(spot)

	if err != nil {
		t.Errorf("test is nil")
	}
}
