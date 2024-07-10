package depth

import (
	"context"

	"github.com/adshao/go-binance/v2"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func Init(d *depth_types.Depth, client *binance.Client, limit int) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(d.Symbol())).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	d.ClearBids()
	for _, bid := range res.Bids {
		price, quantity, _ := bid.Parse()
		d.SetBid(price, quantity)
	}
	d.ClearAsks()
	for _, ask := range res.Asks {
		price, quantity, _ := ask.Parse()
		d.SetAsk(price, quantity)
	}
	d.LastUpdateID = res.LastUpdateID
	return nil
}
