package handlers

import (
	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
	// trades_interface "github.com/fr0ster/go-trading-utils/interfaces/trade"
)

func GetAggTradesUpdateGuard(trade *trade_types.AggTrades, source chan *binance.WsAggTradeEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			trade.Lock() // Locking the depths
			trade.Update(&trade_types.AggTrade{
				AggTradeID:   event.AggTradeID,
				Price:        event.Price,
				Quantity:     event.Quantity,
				FirstTradeID: event.FirstBreakdownTradeID,
				LastTradeID:  event.LastBreakdownTradeID,
				Timestamp:    event.TradeTime,
				IsBuyerMaker: event.IsBuyerMaker,
				// IsBestPriceMatch:   event.IsBestPriceMatch,
			})
			trade.Unlock() // Unlocking the depths
			out <- true
			source <- event
		}
	}()
	return
}
