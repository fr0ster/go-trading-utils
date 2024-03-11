package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
)

func GetBookTickerMapUpdateHandler() (wsHandler binance.WsBookTickerHandler, bookTickerEventChan chan bool) {
	bookTickerEventChan = make(chan bool)
	wsHandler = func(event *binance.WsBookTickerEvent) {
		bookTickerUpdate := binance.BookTicker{
			Symbol:      event.Symbol,
			BidPrice:    event.BestBidPrice,
			BidQuantity: event.BestBidQty,
			AskPrice:    event.BestAskPrice,
			AskQuantity: event.BestAskQty,
		}

		info.SetBookTicker(info.SymbolName(event.Symbol), bookTickerUpdate)
		bookTickerEventChan <- true
	}
	return
}
