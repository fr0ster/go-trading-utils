package depth_test

import (
	"os"
	"testing"

	depth_analyzer "github.com/fr0ster/go-trading-utils/analyzer/depth"
	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	analyzer_interface "github.com/fr0ster/go-trading-utils/interfaces/analyzer"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func TestDepthAnalyzerLoad(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	limit := 10
	degree := 3
	symbol := "BTCUSDT"
	depth := depth_types.NewDepth(degree, symbol)
	spot_depth.SpotDepthInit(depth, api_key, secret_key, symbol, limit, UseTestnet)

	da := depth_analyzer.New(degree)
	da.Update(depth)

	test := func(da analyzer_interface.DepthAnalyzer) {
		if da == nil {
			t.Errorf("DepthAnalyzerLoad returned an empty map")
		}
		levels := da.GetLevels()
		if levels == nil {
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
