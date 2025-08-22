package depths_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/depths"
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
	client := testutil.SpotClient(t)
	depths := depth_types.New(
		degree,
		symbol,
		spot_depth.DepthStreamCreator(
			spot_depth.CallBackCreator(),
			spot_depth.WsErrorHandlerCreator()),
		spot_depth.InitCreator(depth_types.DepthAPILimit10, client))
	assert.NotNil(t, depths)
	depths.StreamStart()
	depths.ResetEvent(fmt.Errorf("test"))
	fmt.Println("test pass")
	time.Sleep(3 * time.Second)
	close(quit)
}
