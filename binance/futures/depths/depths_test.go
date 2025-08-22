package depths_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/depths"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
)

func TestEvents(t *testing.T) {
	t.Parallel()
	var (
		quit = make(chan struct{})
	)
	symbol := "BTCUSDT"
	degree := 3
	t.Log("TestEvents")
	client := testutil.FuturesClient(t)
	depths := depth_types.New(
		degree,
		symbol,
		futures_depth.DepthStreamCreator(
			depth_types.DepthStreamLevel20,
			depth_types.DepthStreamRate100ms,
			futures_depth.CallBackCreator(),
			futures_depth.WsErrorHandlerCreator()),
		futures_depth.InitCreator(depth_types.DepthAPILimit10, client))
	assert.NotNil(t, depths)
	depths.StreamStart()
	depths.ResetEvent(fmt.Errorf("test"))
	fmt.Println("test pass")
	time.Sleep(3 * time.Second)
	close(quit)
}
