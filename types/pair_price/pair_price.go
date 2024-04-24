package pair_price

import (
	"github.com/adshao/go-binance/v2/common"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	PairPrice struct {
		Price    float64
		Quantity float64
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *PairPrice) Less(than btree.Item) bool {
	return i.Price < than.(*PairPrice).Price
}

func (i *PairPrice) Equal(than btree.Item) bool {
	return i.Price == than.(*PairPrice).Price
}

func (i *PairPrice) Parse(a common.PriceLevel) {
	i.Price, i.Quantity, _ = a.Parse()
}

func Binance2PairPrice(binanceDepth interface{}) (*PairPrice, error) {
	var val PairPrice
	err := copier.Copy(&val, binanceDepth)
	if err != nil {
		return nil, err
	}
	return &val, nil
}
