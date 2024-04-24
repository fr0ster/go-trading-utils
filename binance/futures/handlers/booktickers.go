package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBookTickersUpdateGuard(bookTickers *bookticker_types.BookTickers, source chan *futures.WsBookTickerEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			bookTickerUpdate := &bookticker_types.BookTicker{
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
			source <- event
		}
	}()
	return
}
