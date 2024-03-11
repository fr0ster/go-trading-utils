package info

import "github.com/adshao/go-binance/v2"

type (
	BookTickerMapType map[SymbolType]binance.BookTicker
	BookTickerItem    struct {
		Symbol      SymbolType
		BidPrice    PriceType
		BidQuantity PriceType
		AskPrice    PriceType
		AskQuantity PriceType
	}
	PriceType  float64
	SymbolType string
)
