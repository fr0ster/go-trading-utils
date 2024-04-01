package handlers

import (
	"github.com/adshao/go-binance/v2"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func GetDepthsUpdateGuard(depths *depth_types.Depth, source chan *binance.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			depths.Lock() // Locking the depths
			if int64(depths.BidLastUpdateID)+1 < event.FirstUpdateID {
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						continue
					}
					// depths.Lock()
					depths.SetBid(price, quantity)
					depths.RestrictAsk(price)
					depths.RestrictBid(price)
					// depths.Unlock()
				}
			}
			if int64(depths.AskLastUpdateID)+1 < event.FirstUpdateID {
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						continue
					}
					// depths.Lock()
					depths.SetAsk(price, quantity)
					depths.RestrictAsk(price)
					depths.RestrictBid(price)
					// depths.Unlock()
				}
			}
			depths.Unlock() // Unlocking the depths
			out <- true
		}
	}()
	return
}
