package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetBookTickersUpdateHandler() (wsHandler binance.WsBookTickerHandler, bookTickerEventChan chan bool) {
	bookTickerEventChan = make(chan bool)
	wsHandler = func(event *binance.WsBookTickerEvent) {
		bookTickerUpdate := markets.BookTickerItem{
			Symbol:      markets.SymbolType(event.Symbol),
			BidPrice:    markets.PriceType(utils.ConvStrToFloat64(event.BestBidPrice)),
			BidQuantity: markets.PriceType(utils.ConvStrToFloat64(event.BestBidQty)),
			AskPrice:    markets.PriceType(utils.ConvStrToFloat64(event.BestAskPrice)),
			AskQuantity: markets.PriceType(utils.ConvStrToFloat64(event.BestAskQty)),
		}

		markets.SetBookTicker(bookTickerUpdate)
		bookTickerEventChan <- true
	}
	return
}
