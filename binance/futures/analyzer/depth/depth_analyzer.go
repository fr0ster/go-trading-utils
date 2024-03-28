package depth_analyzer

import (
	futuresDepth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	depth_analyzer_types "github.com/fr0ster/go-trading-utils/types/analyzer/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func Init(a *depth_analyzer_types.DepthAnalyzer, api_key, secret_key, symbolname string, rounded, limits int, UseTestnet bool) error {
	depth := depth_types.NewDepth(3, symbolname)
	err := futuresDepth.FuturesDepthInit(depth, api_key, secret_key, symbolname, limits, UseTestnet)
	if err != nil {
		return err
	}
	return a.Update(depth)
}
