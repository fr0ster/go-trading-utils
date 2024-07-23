package symbol

import (
	"github.com/adshao/go-binance/v2"
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
