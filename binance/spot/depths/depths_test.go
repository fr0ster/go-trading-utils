package depths_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/stretchr/testify/assert"

	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/depths"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
)

func TestEvents(t *testing.T) {
	var (
		quit = make(chan struct{})
	)
	symbol := "BTCUSDT"
	degree := 3
	t.Log("TestEvents")
	api_key := os.Getenv("SPOT_TEST_BINANCE_API_KEY")
	api_secret := os.Getenv("SPOT_TEST_BINANCE_SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, api_secret)
	depths := depth_types.New(
		degree,
		symbol,
		spot_depth.DepthStreamCreator(
			spot_depth.CallBackCreator(),
			spot_depth.WsErrorHandlerCreator()),
		spot_depth.InitCreator(depth_types.DepthAPILimit10, client))
	assert.NotNil(t, depths)
	depths.DepthEventStart()
	depths.ResetEvent(fmt.Errorf("test"))
	fmt.Println("test pass")
	time.Sleep(3 * time.Second)
	close(quit)
}
