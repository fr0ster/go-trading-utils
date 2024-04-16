package symbol

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	FuturesSymbol futures.Symbol
)

func (s *FuturesSymbol) Less(than btree.Item) bool {
	return s.Symbol < than.(*FuturesSymbol).Symbol
}

func (s *FuturesSymbol) Equal(than btree.Item) bool {
	return s.Symbol == than.(*FuturesSymbol).Symbol
}

func (s *FuturesSymbol) GetSymbol() string {
	return s.Symbol
}

func (s *FuturesSymbol) GetFilter(filterType string) interface{} {
	for _, filter := range s.Filters {
		if _, exists := filter["filterType"]; exists && filter["filterType"] == filterType {
			return &filter
		}
	}
	return nil
}

func (s *FuturesSymbol) GetFuturesSymbols() (*binance.Symbol, error) {
	var outSymbol binance.Symbol
	err := copier.Copy(&outSymbol, s)
	if err != nil {
		return nil, err
	}
	return &outSymbol, nil
}

func (s *FuturesSymbol) GetFuturesSymbol() (*futures.Symbol, error) {
	var outSymbol futures.Symbol
	err := copier.Copy(&outSymbol, s)
	if err != nil {
		return nil, err
	}
	return &outSymbol, nil
}

// func NewFuturesSymbol(symbol interface{}) *FuturesSymbols {
// 	val, _ := Binance2FuturesSymbol(symbol)
// 	return val
// }

// func Binance2FuturesSymbol(binanceSymbol interface{}) (*FuturesSymbols, error) {
// 	var symbol FuturesSymbols
// 	err := copier.Copy(&symbol, binanceSymbol)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &symbol, nil
// }
