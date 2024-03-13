package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetFilledOrdersGuard(source chan *binance.WsUserDataEvent) (out chan *binance.WsUserDataEvent) {
	out = make(chan *binance.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == binance.UserDataEventTypeExecutionReport &&
				(event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
					event.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled)) {
				out <- event
			}
		}
	}()
	return
}

func GetBalancesUpdateGuard(balances *markets.BalanceBTree, source chan *binance.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.WsAccountUpdates {
				accountUpdate := markets.BalanceItemType{
					Asset:  item.Asset,
					Free:   utils.ConvStrToFloat64(item.Free),
					Locked: utils.ConvStrToFloat64(item.Locked),
				}
				balances.SetItem(accountUpdate)
			}
			out <- true
		}
	}()
	return
}

func GetBookTickersUpdateGuard(bookTickers *markets.BookTickerBTree, source chan *binance.WsBookTickerEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			bookTickerUpdate := markets.BookTickerItemType{
				Symbol:      markets.SymbolType(event.Symbol),
				BidPrice:    markets.PriceType(utils.ConvStrToFloat64(event.BestBidPrice)),
				BidQuantity: markets.PriceType(utils.ConvStrToFloat64(event.BestBidQty)),
				AskPrice:    markets.PriceType(utils.ConvStrToFloat64(event.BestAskPrice)),
				AskQuantity: markets.PriceType(utils.ConvStrToFloat64(event.BestAskQty)),
			}
			bookTickers.SetItem(bookTickerUpdate)
			out <- true
		}
	}()
	return out
}

func GetDepthsUpdateGuard(depths *markets.DepthBTree, source chan *binance.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
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
			out <- true
		}
	}()
	return
}
