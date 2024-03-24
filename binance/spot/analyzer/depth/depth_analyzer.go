package depth_analyzer

import (
	spotDepth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	analyzer_types "github.com/fr0ster/go-trading-utils/types/analyzer"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func Init(a *analyzer_types.DepthAnalyzer, api_key, secret_key, symbolname string, rounded, limits int, UseTestnet bool) error {
	// client := binance.NewClient(api_key, secret_key)
	// depth, err := client.NewDepthService().Symbol(symbolname).Limit(limits).Do(context.Background())
	// if err != nil {
	// 	return err
	// }
	depth := depth_types.NewDepth(3, symbolname)
	spotDepth.SpotDepthInit(depth, api_key, secret_key, symbolname, limits, UseTestnet)
	a.Update(depth)
	// for _, bid := range depth.Bids {
	// 	price, quantity, _ := bid.Parse()
	// 	a.Set(types.DepthSideBid, &types.DepthLevels{
	// 		Price:    utils.RoundToDecimalPlace(price, rounded),
	// 		Quantity: quantity,
	// 	})
	// }
	// for _, ask := range depth.Asks {
	// 	price, quantity, _ := ask.Parse()
	// 	a.Set(types.DepthSideAsk, &types.DepthLevels{
	// 		Price:    utils.RoundToDecimalPlace(price, rounded),
	// 		Quantity: quantity,
	// 	})
	// }
	return nil
}
