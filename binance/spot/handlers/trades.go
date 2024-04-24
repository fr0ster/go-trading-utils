package handlers

import (
	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
	// trades_interface "github.com/fr0ster/go-trading-utils/interfaces/trade"
)

func GetTradesUpdateGuard(trade *trade_types.Trades, source chan *binance.WsTradeEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			trade.Lock() // Locking the depths
			trade.Update(&trade_types.Trade{
				ID:       event.TradeID,
				Price:    event.Price,
				Quantity: event.Quantity,
				// QuoteQuantity: event.QuoteQuantity,
				Time:         event.Time,
				IsBuyerMaker: event.IsBuyerMaker,
				// IsBestMatch:   event.IsBestMatch,
				// IsIsolated:    event.IsIsolated,
			})
			trade.Unlock() // Unlocking the depths
			out <- true
		}
	}()
	return
}
