package depth_test

import (
	"os"
	"testing"

	depth_analyzer "github.com/fr0ster/go-trading-utils/analyzer/depth"
	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	analyzer_interface "github.com/fr0ster/go-trading-utils/interfaces/analyzer"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
)

func TestDepthAnalyzerLoad(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	limit := 10
	degree := 3
	round := 2
	symbol := "BTCUSDT"
	depth := spot_depth.New(degree, round, limit, symbol)
	depth.Init(api_key, secret_key, symbol, UseTestnet)

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
