package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
)

var (
	bookTickerMap     = make(BookTickerMapType)
	mu_bookticker_map sync.Mutex
)

func BookTickerMapMutexLock() {
	mu_bookticker_map.Lock()
}

func BookTickerMapMutexUnlock() {
	mu_bookticker_map.Unlock()
}

func InitPricesMap(client *binance.Client, symbolname string) (err error) {
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	for _, bookTicker := range bookTickerList {
		bookTickerMap[SymbolType(bookTicker.Symbol)] = *bookTicker
	}
	return nil
}

func GetBookTickerMap() BookTickerMapType {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	return bookTickerMap
}

func GetBookTickerMapItem(symbolname SymbolType) binance.BookTicker {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	return bookTickerMap[symbolname]
}

func SetBookTickerMapItem(symbolname SymbolType, bookticker binance.BookTicker) {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	bookTickerMap[symbolname] = bookticker
}

func ShowBookTickerMap() {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	for k, v := range bookTickerMap {
		println(
			"Symbol:", k,
			"BidPrice:", v.BidPrice,
			"BidQuantity:", v.BidQuantity,
			"AskPrice:", v.AskPrice,
			"AskQuantity:", v.AskQuantity)
	}
}
