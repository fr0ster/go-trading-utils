package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetDepthUpdateHandler() (wsHandler binance.WsDepthHandler, depthChan chan bool) {
	depthChan = make(chan bool)
	wsHandler = func(event *binance.WsDepthEvent) {
		for _, bid := range event.Bids {
			value, exists := info.GetDepthMapItem(info.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
				value.BidQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.BidLastUpdateID = event.LastUpdateID
			} else {
				info.SetDepthMapItem(info.Price(utils.ConvStrToFloat64(bid.Price)),
					info.DepthRecord{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     0,
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
					})
			}
		}

		for _, bid := range event.Asks {
			value, exists := info.GetDepthMapItem(info.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
				value.AskQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.AskLastUpdateID = event.LastUpdateID
			} else {
				info.SetDepthMapItem(info.Price(utils.ConvStrToFloat64(bid.Price)),
					info.DepthRecord{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					})
			}
		}
		depthChan <- true
	}
	return
}

func GetDepthUpdateHandlerTree() (wsHandler binance.WsDepthHandler, depthChan chan bool) {
	depthChan = make(chan bool)
	wsHandler = func(event *binance.WsDepthEvent) {
		for _, bid := range event.Bids {
			value, exists := info.GetDepthTreeItem(info.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
				value.BidQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.BidLastUpdateID = event.LastUpdateID
			} else {
				value =
					info.DepthRecord{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
			info.SetDepthTreeItem(value)
		}

		for _, bid := range event.Asks {
			value, exists := info.GetDepthTreeItem(info.Price(utils.ConvStrToFloat64(bid.Price)))
			if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
				value.AskQuantity += info.Price(utils.ConvStrToFloat64(bid.Quantity))
				value.AskLastUpdateID = event.LastUpdateID
			} else {
				value =
					info.DepthRecord{
						Price:           info.Price(utils.ConvStrToFloat64(bid.Price)),
						AskLastUpdateID: event.LastUpdateID,
						AskQuantity:     info.Price(utils.ConvStrToFloat64(bid.Quantity)),
						BidLastUpdateID: event.LastUpdateID,
						BidQuantity:     0,
					}
			}
			info.SetDepthTreeItem(value)
		}
		depthChan <- true
	}
	return
}
