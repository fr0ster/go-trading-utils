package pair_price

import (
	"github.com/adshao/go-binance/v2/common"
	"github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	PairPrice struct {
		Price    types.PriceType
		Quantity types.QuantityType
	}
	PairDelta struct {
		Price   types.PriceType
		Percent types.QuantityType
	}
	AskBid struct {
		Ask *PairDelta
		Bid *PairDelta
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
	price, quantity, _ := a.Parse()
	i.Price = types.PriceType(price)
	i.Quantity = types.QuantityType(quantity)
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
