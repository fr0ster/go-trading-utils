package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetFilledOrdersGuard() (executeOrderChan chan *binance.WsUserDataEvent) {
	executeOrderChan = make(chan *binance.WsUserDataEvent, 1)
	go func() {
		userDataChannel, err := GetUserDataChannel()
		if !err {
			return
		}
		for {
			event := <-userDataChannel
			if event.Event == binance.UserDataEventTypeExecutionReport &&
				(event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
					event.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled)) {
				executeOrderChan <- event
			}
		}
	}()
	return
}

func GetBalancesUpdateGuard() (accountEventChan chan bool) {
	accountEventChan = make(chan bool)
	go func() {
		accountChan, res := GetUserDataChannel()
		if !res {
			return
		}
		for {
			event := <-accountChan
			for _, item := range event.AccountUpdate.WsAccountUpdates {
				accountUpdate := markets.BalanceItemType{
					Asset:  item.Asset,
					Free:   utils.ConvStrToFloat64(item.Free),
					Locked: utils.ConvStrToFloat64(item.Locked),
				}
				markets.GetBalancesTree().ReplaceOrInsert(accountUpdate)
			}
			accountEventChan <- true
		}
	}()
	return
}

func GetBookTickersUpdateGuard(bookTickers *markets.BookTickerBTree) (bookTickerEventChan chan bool) {
	bookTickerEventChan = make(chan bool)
	go func() {
		// bookTickerChan, res := GetBookTickerChannel()
		// if !res {
		// 	return
		// }
		for {
			// event := <-bookTickerChan
			// value, exists := bookTickers.GetItem(markets.SymbolType(event.Symbol))
			// bookTickerUpdate := markets.BookTickerItemType{
			// 	Symbol:      markets.SymbolType(event.Symbol),
			// 	BidPrice:    markets.PriceType(utils.ConvStrToFloat64(event.BestBidPrice)),
			// 	BidQuantity: markets.PriceType(utils.ConvStrToFloat64(event.BestBidQty)),
			// 	AskPrice:    markets.PriceType(utils.ConvStrToFloat64(event.BestAskPrice)),
			// 	AskQuantity: markets.PriceType(utils.ConvStrToFloat64(event.BestAskQty)),
			// }
			// markets.SetBookTicker(bookTickerUpdate)
			bookTickerEventChan <- true
		}
	}()
	return bookTickerEventChan
}

func GetDepthsUpdateGuard(depths *markets.DepthBTree) (depthBoolChan chan bool) {
	depthBoolChan = make(chan bool)
	go func() {
		depthChan, res := GetDepthChannel()
		if !res {
			return
		}
		for {
			event := <-depthChan
			for _, bid := range event.Bids {
				value, exists := depths.GetItem(markets.Price(utils.ConvStrToFloat64(bid.Price)))
				if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
					value.BidQuantity += markets.Price(utils.ConvStrToFloat64(bid.Quantity))
					value.BidLastUpdateID = event.LastUpdateID
				} else {
					value =
						&markets.DepthItemType{
							Price:           markets.Price(utils.ConvStrToFloat64(bid.Price)),
							AskLastUpdateID: event.LastUpdateID,
							AskQuantity:     markets.Price(utils.ConvStrToFloat64(bid.Quantity)),
							BidLastUpdateID: event.LastUpdateID,
							BidQuantity:     0,
						}
				}
				depths.SetItem(*value)
			}

			for _, bid := range event.Asks {
				value, exists := depths.GetItem(markets.Price(utils.ConvStrToFloat64(bid.Price)))
				if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
					value.AskQuantity += markets.Price(utils.ConvStrToFloat64(bid.Quantity))
					value.AskLastUpdateID = event.LastUpdateID
				} else {
					value =
						&markets.DepthItemType{
							Price:           markets.Price(utils.ConvStrToFloat64(bid.Price)),
							AskLastUpdateID: event.LastUpdateID,
							AskQuantity:     markets.Price(utils.ConvStrToFloat64(bid.Quantity)),
							BidLastUpdateID: event.LastUpdateID,
							BidQuantity:     0,
						}
				}
				depths.SetItem(*value)
			}
			depthBoolChan <- true
		}
	}()
	return
}
