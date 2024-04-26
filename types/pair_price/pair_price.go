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
	AskBid struct {
		Ask *PairPrice
		Bid *PairPrice
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

func Binance2PairPrice(binancePairPrice interface{}) (*PairPrice, error) {
	var val PairPrice
	err := copier.Copy(&val, binancePairPrice)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (ab *AskBid) Less(than btree.Item) bool {
	return ab.Ask.Price < than.(*AskBid).Ask.Price
}

func (ab *AskBid) Equal(than btree.Item) bool {
	return ab.Ask.Price == than.(*AskBid).Ask.Price
}
