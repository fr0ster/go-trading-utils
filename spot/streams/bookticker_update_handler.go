package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetBookTickersUpdateHandler() (wsHandler binance.WsBookTickerHandler, bookTickerEventChan chan bool) {
	bookTickerEventChan = make(chan bool)
	wsHandler = func(event *binance.WsBookTickerEvent) {
		bookTickerUpdate := info.BookTickerItem{
			Symbol:      info.SymbolType(event.Symbol),
			BidPrice:    info.PriceType(utils.ConvStrToFloat64(event.BestBidPrice)),
			BidQuantity: info.PriceType(utils.ConvStrToFloat64(event.BestBidQty)),
			AskPrice:    info.PriceType(utils.ConvStrToFloat64(event.BestAskPrice)),
			AskQuantity: info.PriceType(utils.ConvStrToFloat64(event.BestAskQty)),
		}

		info.SetBookTicker(bookTickerUpdate)
		bookTickerEventChan <- true
	}
	return
}
