package depth_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"

	"github.com/stretchr/testify/assert"
)

func TestInitDepthTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	futures := futures.NewClient(api_key, secret_key)

	// Add more test cases here
	testDepthTree := depth_types.New(3, "SUSHIUSDT", false, 10, 100, depth_types.DepthStreamRate100ms)
	err := futures_depth.Init(testDepthTree, futures)
	assert.NoError(t, err)
}
