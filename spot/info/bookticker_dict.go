package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
)

var (
	BookTickerMap     = make(BookTickerMapType)
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
		BookTickerMap[SymbolName(bookTicker.Symbol)] = *bookTicker
	}
	return nil
}

func GetBookTickerMap() BookTickerMapType {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	return BookTickerMap
}

func GetBookTicker(symbolname SymbolName) binance.BookTicker {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	return BookTickerMap[symbolname]
}

func SetBookTicker(symbolname SymbolName, bookticker binance.BookTicker) {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	BookTickerMap[symbolname] = bookticker
}

func ShowBookTickerMap() {
	mu_bookticker_map.Lock()
	defer mu_bookticker_map.Unlock()
	for k, v := range BookTickerMap {
		println(
			"Symbol:", k,
			"BidPrice:", v.BidPrice,
			"BidQuantity:", v.BidQuantity,
			"AskPrice:", v.AskPrice,
			"AskQuantity:", v.AskQuantity)
	}
}
