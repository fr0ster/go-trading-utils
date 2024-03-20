package handlers

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/balances"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	bookticker_interface "github.com/fr0ster/go-trading-utils/interfaces/bookticker"
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

func GetBalancesUpdateGuard(bt *balances.BalanceBTree, source chan *binance.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.WsAccountUpdates {
				balanceUpdate := balances.BalanceItemType{
					Asset:  balances.AssetType(item.Asset),
					Free:   utils.ConvStrToFloat64(item.Free),
					Locked: utils.ConvStrToFloat64(item.Locked),
				}
				bt.Lock()
				bt.SetItem(balanceUpdate)
				bt.Unlock()
			}
			out <- true
		}
	}()
	return
}

func GetBookTickersUpdateGuard(bookTickers *bookticker.BookTickerBTree, source chan *binance.WsBookTickerEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			bookTickerUpdate := bookticker_interface.BookTickerItem{
				Symbol:      event.Symbol,
				BidPrice:    utils.ConvStrToFloat64(event.BestBidPrice),
				BidQuantity: utils.ConvStrToFloat64(event.BestBidQty),
				AskPrice:    utils.ConvStrToFloat64(event.BestAskPrice),
				AskQuantity: utils.ConvStrToFloat64(event.BestAskQty),
			}
			bookTickers.Lock()
			bookTickers.Set(bookTickerUpdate)
			bookTickers.Unlock()
			out <- true
		}
	}()
	return out
}

func GetDepthsUpdateGuard(depths *depth.Depth, source chan *binance.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			if int64(depths.BidLastUpdateID)+1 < event.FirstUpdateID {
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						continue
					}
					depths.Lock()
					depths.UpdateBid(price, quantity)
					depths.Unlock()
				}
			}
			if int64(depths.AskLastUpdateID)+1 < event.FirstUpdateID {
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						continue
					}
					depths.Lock()
					depths.UpdateAsk(price, quantity)
					depths.Unlock()
				}
			}
			out <- true
		}
	}()
	return
}
