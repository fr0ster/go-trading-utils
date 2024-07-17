package depth

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/fr0ster/go-trading-utils/types/depth/types"
)

func Init(d *depth_types.Depth, client *futures.Client) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(d.Symbol())).
			Limit(int(d.GetLimitDepth())).
			Do(context.Background())
	if err != nil {
		return err
	}
	d.ClearBids()
	for _, bid := range res.Bids {
		price, quantity, _ := bid.Parse()
		d.UpdateBid(types.PriceType(price), types.QuantityType(quantity))
	}
	d.ClearAsks()
	for _, ask := range res.Asks {
		price, quantity, _ := ask.Parse()
		d.UpdateAsk(types.PriceType(price), types.QuantityType(quantity))
	}
	d.LastUpdateID = res.LastUpdateID
	return nil
}
