package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetDepthsUpdateHandler() (wsHandler binance.WsDepthHandler, depthChan chan bool) {
	depthChan = make(chan bool)
	wsHandler = func(event *binance.WsDepthEvent) {
		for _, bid := range event.Bids {
			value, exists := markets.GetDepth(markets.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
				value.BidQuantity += markets.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.BidLastUpdateID = event.LastUpdateID
			} else {
				value =
					markets.DepthItem{
						Price:           markets.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     markets.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
			markets.SetDepth(value)
		}

		for _, bid := range event.Asks {
			value, exists := markets.GetDepth(markets.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
				value.AskQuantity += markets.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.AskLastUpdateID = event.LastUpdateID
			} else {
				value =
					markets.DepthItem{
						Price:           markets.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     markets.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
			markets.SetDepth(value)
		}
		depthChan <- true
	}
	return
}
