package handlers

import (
	"github.com/adshao/go-binance/v2"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func GetDepthsUpdateGuard(depths *depth_types.Depth, source chan *binance.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		var res bool = false
		for {
			event := <-source
			// Checking if the event is not outdated
			if event.LastUpdateID <= depths.LastUpdateID {
				continue
			}
			depths.Lock() // Locking the depths
			if (int64(depths.LastUpdateID)+1 > event.FirstUpdateID && int64(depths.LastUpdateID)+1 < event.LastUpdateID) ||
				int64(depths.LastUpdateID)+1 < event.LastUpdateID {
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						continue
					}
					res = res || depths.UpdateBid(price, quantity)
				}
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						continue
					}
					res = res || depths.UpdateAsk(price, quantity)
				}
				depths.LastUpdateID = event.LastUpdateID
			}
			depths.Unlock() // Unlocking the depths
			out <- res
		}
	}()
	return
}
