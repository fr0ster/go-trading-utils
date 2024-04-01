package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func GetDepthsUpdateGuard(depths *depth_types.Depth, source chan *futures.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			// Checking if the event is not outdated
			if event.LastUpdateID <= depths.LastUpdateID {
				continue
			}
			depths.Lock() // Locking the depths
			if int64(depths.LastUpdateID)+1 > event.FirstUpdateID && int64(depths.LastUpdateID)+1 < event.LastUpdateID {
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						continue
					}
					depths.Lock()
					depths.SetBid(price, quantity)
					depths.Unlock()
				}
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						continue
					}
					depths.Lock()
					depths.SetAsk(price, quantity)
					depths.Unlock()
				}
			}
			out <- true
		}
	}()
	return
}
