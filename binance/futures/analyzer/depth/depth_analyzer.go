package depth_analyzer

import (
	"github.com/adshao/go-binance/v2/futures"
	futuresDepth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	depth_analyzer_types "github.com/fr0ster/go-trading-utils/types/analyzer/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func Init(a *depth_analyzer_types.DepthAnalyzer, client *futures.Client, symbolname string, rounded, limits int, UseTestnet bool) error {
	depth := depth_types.NewDepth(a.Degree, symbolname)
	err := futuresDepth.Init(depth, client, limits)
	if err != nil {
		return err
	}
	return a.Update(depth)
}
