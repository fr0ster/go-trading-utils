package streams

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/types"
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

func GetBalancesUpdateGuard(balances *markets.BalanceBTree, source chan *futures.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.Balances {
				balanceUpdate := markets.BalanceItemType{
					Asset:              markets.AssetType(item.Asset),
					Balance:            utils.ConvStrToFloat64(item.Balance),
					CrossWalletBalance: utils.ConvStrToFloat64(item.CrossWalletBalance),
					ChangeBalance:      utils.ConvStrToFloat64(item.ChangeBalance),
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

func GetBookTickersUpdateGuard(bookTickers *markets.BookTickerBTree, source chan *futures.WsBookTickerEvent) (out chan bool) {
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

func GetDepthsUpdateGuard(depths *depth.DepthBTree, source chan *futures.WsDepthEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, bid := range event.Bids {
				value, exists := depths.GetItem(types.Price(utils.ConvStrToFloat64(bid.Price)))
				if exists && value.BidLastUpdateID+1 > event.FirstUpdateID {
					value.BidQuantity += types.Price(utils.ConvStrToFloat64(bid.Quantity))
					value.BidLastUpdateID = event.LastUpdateID
				} else {
					value =
						&depth_interface.DepthItemType{
							Price:           types.Price(utils.ConvStrToFloat64(bid.Price)),
							AskLastUpdateID: event.LastUpdateID,
							AskQuantity:     types.Price(utils.ConvStrToFloat64(bid.Quantity)),
							BidLastUpdateID: event.LastUpdateID,
							BidQuantity:     0,
						}
				}
				depths.Lock()
				depths.SetItem(*value)
				depths.Unlock()
			}

			for _, bid := range event.Asks {
				value, exists := depths.GetItem(types.Price(utils.ConvStrToFloat64(bid.Price)))
				if exists && value.AskLastUpdateID+1 > event.FirstUpdateID {
					value.AskQuantity += types.Price(utils.ConvStrToFloat64(bid.Quantity))
					value.AskLastUpdateID = event.LastUpdateID
				} else {
					value =
						&depth_interface.DepthItemType{
							Price:           types.Price(utils.ConvStrToFloat64(bid.Price)),
							AskLastUpdateID: event.LastUpdateID,
							AskQuantity:     types.Price(utils.ConvStrToFloat64(bid.Quantity)),
							BidLastUpdateID: event.LastUpdateID,
							BidQuantity:     0,
						}
				}
				depths.Lock()
				depths.SetItem(*value)
				depths.Unlock()
			}
			out <- true
		}
	}()
	return
}
