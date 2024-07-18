package handlers

import (
	"github.com/adshao/go-binance/v2"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func GetDepthsUpdateGuard(depths *depth_types.Depths, source chan *binance.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		var res bool = false
		for {
			event := <-source
			// Checking if the event is not outdated
			if event.LastUpdateID <= depths.LastUpdateID {
				continue
			}
			if depths.LastUpdateID == 0 {
				depths.LastUpdateID = event.FirstUpdateID + 1
			}
			if event.FirstUpdateID <= int64(depths.LastUpdateID)+1 && event.LastUpdateID >= int64(depths.LastUpdateID)+1 {
				depths.Lock() // Locking the depths
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						continue
					}
					res = depths.GetBids().Update(types.NewBid(types.PriceType(price), types.QuantityType(quantity))) || res
				}
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						continue
					}
					res = depths.GetAsks().Update(types.NewAsk(types.PriceType(price), types.QuantityType(quantity))) || res
				}
				depths.LastUpdateID = event.LastUpdateID
				depths.Unlock() // Unlocking the depths
			}
			if res {
				out <- res
			}
		}
	}()
	return
}
