package handlers

import (
	"github.com/adshao/go-binance/v2"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBookTickersUpdateGuard(bookTickers *bookticker_types.BookTickers, source chan *binance.WsBookTickerEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			currentBookTicker := bookTickers.Get(event.Symbol)
			if currentBookTicker != nil &&
				currentBookTicker.(*bookticker_types.BookTicker).UpdateID >= event.UpdateID {
				continue
			}
			bookTickerUpdate := &bookticker_types.BookTicker{
				UpdateID:    event.UpdateID,
				Symbol:      event.Symbol,
				BidPrice:    utils.ConvStrToFloat64(event.BestBidPrice),
				BidQuantity: utils.ConvStrToFloat64(event.BestBidQty),
				AskPrice:    utils.ConvStrToFloat64(event.BestAskPrice),
				AskQuantity: utils.ConvStrToFloat64(event.BestAskQty),
			}
			bookTickers.Lock() // Locking the bookTickers
			bookTickers.Set(bookTickerUpdate)
			bookTickers.Unlock() // Unlocking the bookTickers
			out <- true
		}
	}()
	return
}
