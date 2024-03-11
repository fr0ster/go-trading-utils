package info

import "github.com/adshao/go-binance/v2"

type (
	BookTickerMapType map[SymbolName]binance.BookTicker
	BookTickerItem    struct {
		Symbol      SymbolName
		BidPrice    SymbolPrice
		BidQuantity SymbolPrice
		AskPrice    SymbolPrice
		AskQuantity SymbolPrice
	}
	SymbolPrice float64
	SymbolName  string
)
