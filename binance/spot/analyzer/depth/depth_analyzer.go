package depth_analyzer

import (
	"github.com/adshao/go-binance/v2"
	spotDepth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	depth_analyzer_types "github.com/fr0ster/go-trading-utils/types/analyzer/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func Init(a *depth_analyzer_types.DepthAnalyzer, client *binance.Client, symbolname string, limits int) error {
	depth := depth_types.New(a.Degree, symbolname)
	err := spotDepth.Init(depth, client, limits)
	if err != nil {
		return err
	}
	return a.Update(depth)
}
