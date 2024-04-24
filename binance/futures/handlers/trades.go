package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func GetAggTradesUpdateGuard(trade *trade_types.AggTrades, source chan *futures.WsAggTradeEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			trade.Lock() // Locking the depths
			trade.Update(&trade_types.AggTrade{
				AggTradeID:   event.AggregateTradeID,
				Price:        event.Price,
				Quantity:     event.Quantity,
				FirstTradeID: event.FirstTradeID,
				LastTradeID:  event.LastTradeID,
				Timestamp:    event.TradeTime,
				// IsBuyerMaker:     event.IsBuyerMaker,
				// IsBestPriceMatch: event.IsBestPriceMatch,
			})
			// trade.Unlock()
			trade.Unlock() // Unlocking the depths
			out <- true
			source <- event
		}
	}()
	return
}
