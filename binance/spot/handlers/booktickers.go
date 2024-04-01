package handlers

import (
	"github.com/adshao/go-binance/v2"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBookTickersUpdateGuard(bookTickers *bookticker_types.BookTickerBTree, source chan *binance.WsBookTickerEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			bookTickers.Lock() // Locking the bookTickers
			bookTickerUpdate := &bookticker_types.BookTickerItem{
				Symbol:      event.Symbol,
				BidPrice:    utils.ConvStrToFloat64(event.BestBidPrice),
				BidQuantity: utils.ConvStrToFloat64(event.BestBidQty),
				AskPrice:    utils.ConvStrToFloat64(event.BestAskPrice),
				AskQuantity: utils.ConvStrToFloat64(event.BestAskQty),
			}
			// bookTickers.Lock()
			bookTickers.Set(bookTickerUpdate)
			bookTickers.Unlock() // Unlocking the bookTickers
			out <- true
		}
	}()
	return
}
