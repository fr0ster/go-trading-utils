package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	"github.com/fr0ster/go-trading-utils/utils"
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
				balanceUpdate := markets.BalanceItemType{
					Asset:  markets.AssetType(item.Asset),
					Free:   utils.ConvStrToFloat64(item.Free),
					Locked: utils.ConvStrToFloat64(item.Locked),
				}
				balances.Lock()
				balances.SetItem(balanceUpdate)
				balances.Unlock()
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
			bookTickers.Lock()
			bookTickers.SetItem(bookTickerUpdate)
			bookTickers.Unlock()
			out <- true
		}
	}()
	return out
}

func GetDepthsUpdateGuard(depths *depth.DepthBTree, source chan *binance.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			if int64(depths.BidLastUpdateID)+1 > event.FirstUpdateID {
				for _, bid := range event.Bids {
					depths.Lock()
					depths.UpdateBid(bid, event.LastUpdateID)
					depths.Unlock()
				}
			}
			if int64(depths.AskLastUpdateID)+1 > event.FirstUpdateID {
				for _, ask := range event.Asks {
					depths.Lock()
					depths.UpdateAsk(ask, event.LastUpdateID)
					depths.Unlock()
				}
			}
			out <- true
		}
	}()
	return
}
