package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/balances"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetFilledOrdersGuard(source chan *futures.WsUserDataEvent) (out chan *futures.WsUserDataEvent) {
	out = make(chan *futures.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
				(event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
					event.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled) {
				out <- event
			}
		}
	}()
	return
}

func GetBalancesUpdateGuard(bt *balances.BalanceBTree, source chan *futures.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.Balances {
				balanceUpdate := balances.BalanceItemType{
					Asset:  balances.AssetType(item.Asset),
					Free:   utils.ConvStrToFloat64(item.Balance) - utils.ConvStrToFloat64(item.ChangeBalance),
					Locked: utils.ConvStrToFloat64(item.ChangeBalance),
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

func GetBookTickersUpdateGuard(bookTickers *bookticker_types.BookTickerBTree, source chan *futures.WsBookTickerEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			bookTickerUpdate := &bookticker_types.BookTickerItem{
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

func GetDepthsUpdateGuard(depths *depth.Depth, source chan *futures.WsDepthEvent) (out chan bool) {
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
					depths.SetBid(price, quantity)
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
					depths.SetAsk(price, quantity)
					depths.Unlock()
				}
			}
			out <- true
		}
	}()
	return
}
