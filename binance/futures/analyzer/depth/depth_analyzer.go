package depth_analyzer

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	types "github.com/fr0ster/go-trading-utils/types"
	depth_analyzer_types "github.com/fr0ster/go-trading-utils/types/analyzer/depth"
)

func Init(a *depth_analyzer_types.DepthAnalyzer, api_key, secret_key, symbolname string, limits int, UseTestnet bool) error {
	client := futures.NewClient(api_key, secret_key)
	depth, err := client.NewDepthService().Symbol(symbolname).Limit(limits).Do(context.Background())
	if err != nil {
		return err
	}
	for _, bid := range depth.Bids {
		price, quantity, _ := bid.Parse()
		a.Set(types.DepthSideBid, &types.DepthLevels{
			Price:    price,
			Quantity: quantity,
		})
	}
	for _, ask := range depth.Asks {
		price, quantity, _ := ask.Parse()
		a.Set(types.DepthSideAsk, &types.DepthLevels{
			Price:    price,
			Quantity: quantity,
		})
	}
	return nil
}
