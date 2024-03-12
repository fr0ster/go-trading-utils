package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetDepthsUpdateHandler() (wsHandler binance.WsDepthHandler, depthChan chan bool) {
	depthChan = make(chan bool)
	wsHandler = func(event *binance.WsDepthEvent) {
		for _, bid := range event.Bids {
			value, exists := info.GetDepth(info.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
				value.BidQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.BidLastUpdateID = event.LastUpdateID
			} else {
				value =
					info.DepthItem{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
			info.SetDepth(value)
		}

		for _, bid := range event.Asks {
			value, exists := info.GetDepth(info.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
				value.AskQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.AskLastUpdateID = event.LastUpdateID
			} else {
				value =
					info.DepthItem{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
			info.SetDepth(value)
		}
		depthChan <- true
	}
	return
}
