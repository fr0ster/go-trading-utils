package streams

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/fr0ster/go-binance-utils/spot/utils"
)

func GetDepthUpdateHandler(mu *sync.Mutex) (wsHandler binance.WsDepthHandler, depthChan chan bool) {
	depthChan = make(chan bool)
	wsHandler = func(event *binance.WsDepthEvent) {
		mu.Lock()
		defer mu.Unlock()
		depthMap := info.GetDepthMap()
		for _, bid := range event.Bids {
			value, exists := (*depthMap)[info.Price(utils.ConvStrToFloat64(bid.Price))]
			if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
				value.BidQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.BidLastUpdateID = event.LastUpdateID
			} else {
				(*depthMap)[info.Price(utils.ConvStrToFloat64(bid.Price))] =
					info.DepthRecord{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     0,
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
					}
			}
		}

		for _, bid := range event.Asks {
			value, exists := (*depthMap)[info.Price(utils.ConvStrToFloat64(bid.Price))]
			if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
				value.AskQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.AskLastUpdateID = event.LastUpdateID
			} else {
				(*depthMap)[info.Price(utils.ConvStrToFloat64(bid.Price))] =
					info.DepthRecord{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
		}
		depthChan <- true
	}
	return
}
