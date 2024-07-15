package depth_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"

	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func TestInitDepthTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	spot := binance.NewClient(api_key, secret_key)

	// Add more test cases here
	testDepthTree := depth_types.New(3, "SUSHIUSDT", false, 10, 100, 1, depth_types.DepthStreamRate100ms)
	err := spot_depth.Init(testDepthTree, spot)
	if err != nil {
		t.Errorf("Failed to initialize depth tree: %v", err)
	}
}
