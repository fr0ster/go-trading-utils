package symbol

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	SpotSymbol binance.Symbol
)

func (s *SpotSymbol) Less(than btree.Item) bool {
	return s.Symbol < than.(*SpotSymbol).Symbol
}

func (s *SpotSymbol) Equal(than btree.Item) bool {
	return s.Symbol == than.(*SpotSymbol).Symbol
}

func (s *SpotSymbol) GetSymbol() string {
	return s.Symbol
}

func (s *SpotSymbol) GetFilter(filterType string) interface{} {
	for _, filter := range s.Filters {
		if _, exists := filter["filterType"]; exists && filter["filterType"] == filterType {
			return &filter
		}
	}
	return nil
}

func (s *SpotSymbol) GetSpotSymbol() (*binance.Symbol, error) {
	var outSymbol binance.Symbol
	err := copier.Copy(&outSymbol, s)
	if err != nil {
		return nil, err
	}
	return &outSymbol, nil
}

func (s *SpotSymbol) GetFuturesSymbol() (*futures.Symbol, error) {
	var outSymbol futures.Symbol
	err := copier.Copy(&outSymbol, s)
	if err != nil {
		return nil, err
	}
	return &outSymbol, nil
}

// func NewSpotSymbol(symbol interface{}) *SpotSymbol {
// 	val, _ := Binance2SpotSymbol(symbol)
// 	return val
// }

// func Binance2SpotSymbol(binanceSymbol interface{}) (*SpotSymbol, error) {
// 	var symbol SpotSymbol
// 	err := copier.Copy(&symbol, binanceSymbol)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &symbol, nil
// }
