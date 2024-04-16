package depth_analyzer_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	depth_analyzer_interface "github.com/fr0ster/go-trading-utils/interfaces/analyzer/depth"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/types"
	depth_analyzer "github.com/fr0ster/go-trading-utils/types/analyzer/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func TestDepthAnalyzerLoad(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	spot := binance.NewClient(api_key, secret_key)
	limit := 10
	degree := 3
	rounded := 2
	bound := 0.5
	symbol := "BTCUSDT"
	depth := depth_types.NewDepth(degree, symbol)
	spot_depth.Init(depth, spot, limit)

	da := depth_analyzer.New(3, rounded, bound)
	da.Update(depth)

	test := func(da depth_analyzer_interface.DepthAnalyzer) {
		if da == nil {
			t.Errorf("DepthAnalyzerLoad returned an empty map")
		}
		askLevels := da.GetLevels(types.DepthSideAsk)
		if askLevels == nil {
			t.Errorf("DepthAnalyzerLoad returned an empty map")
		}
		bidLevels := da.GetLevels(types.DepthSideAsk)
		if bidLevels == nil {
			t.Errorf("DepthAnalyzerLoad returned an empty map")
		}
	}

	test(da)

	test2 := func(ds depth_interface.Depth) {
		if ds == nil {
			t.Errorf("DepthAnalyzerLoad returned an empty map")
		}
	}

	test2(depth)
}
