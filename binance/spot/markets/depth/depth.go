package depth

import (
	"context"

	"github.com/adshao/go-binance/v2"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func Init(d *depth_types.Depths, client *binance.Client) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(d.Symbol())).
			Limit(int(d.GetLimitDepth())).
			Do(context.Background())
	if err != nil {
		return err
	}
	d.GetBids().Clear()
	for _, bid := range res.Bids {
		price, quantity, _ := bid.Parse()
		d.GetBids().Update(types.NewBid(types.PriceType(price), types.QuantityType(quantity)))
	}
	d.GetAsks().Clear()
	for _, ask := range res.Asks {
		price, quantity, _ := ask.Parse()
		d.GetAsks().Update(types.NewAsk(types.PriceType(price), types.QuantityType(quantity)))
	}
	d.LastUpdateID = res.LastUpdateID
	return nil
}
