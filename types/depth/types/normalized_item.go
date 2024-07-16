package types

import (
	"math"

	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	NormalizedItem struct {
		// Службові дані
		exp     int
		roundUp bool
		// Дані по ціні
		price    PriceType
		quantity QuantityType
		minMax   *btree.BTree
		depths   *btree.BTree
	}
)

func (i *NormalizedItem) Less(than btree.Item) bool {
	return i.price < than.(*NormalizedItem).price
}

func (i *NormalizedItem) Equal(than btree.Item) bool {
	return i.price == than.(*NormalizedItem).price
}

func (i *NormalizedItem) GetNormalizedPrice() PriceType {
	return i.price
}

func (i *NormalizedItem) GetQuantity() QuantityType {
	return i.quantity
}

func (i *NormalizedItem) SetQuantity(quantity QuantityType) {
	i.quantity = quantity
}

func (i *NormalizedItem) GetDepth(price PriceType) (depthItem *DepthItem) {
	if i.depths != nil {
		depthItem = i.depths.Get(NewDepthItem(price)).(*DepthItem)
	}
	return
}

func (i *NormalizedItem) SetDepth(depth *DepthItem) {
	i.depths.ReplaceOrInsert(depth)
}

func (i *NormalizedItem) DeleteDepth(depth *DepthItem) {
	i.depths.Delete(depth)
}

func (i *NormalizedItem) GetMinMax(quantity QuantityType) (minMax *QuantityItem) {
	if i.minMax != nil {
		if val := i.minMax.Get(&QuantityItem{quantity: quantity}); val != nil {
			minMax = val.(*QuantityItem)
		}
	}
	return
}

func (i *NormalizedItem) SetMinMax(minMax *QuantityItem) {
	i.minMax.ReplaceOrInsert(minMax)
}

func (i *NormalizedItem) DeleteMinMax(minMax *QuantityItem) {
	i.minMax.Delete(minMax)
}

func getNormalizedPrice(price PriceType, exp int, roundUp bool) (normalizedPrice PriceType, err error) {
	len := int(math.Log10(float64(price))) + 1
	rounded := 0.0
	if len == exp {
		if roundUp {
			normalizedPrice = PriceType(math.Ceil(float64(price)))
		} else {
			normalizedPrice = PriceType(math.Floor(float64(price)))
		}
	} else if len > exp {
		normalized := PriceType(float64(price) * math.Pow(10, float64(-exp)))
		if roundUp {
			rounded = math.Ceil(float64(normalized))
		} else {
			rounded = math.Floor(float64(normalized))
		}
		normalizedPrice = PriceType(utils.RoundToDecimalPlace(rounded*math.Pow(10, float64(exp)), exp))
	} else {
		normalized := float64(price) * math.Pow(10, float64(exp-1))
		if roundUp {
			rounded = math.Ceil(normalized)
		} else {
			rounded = math.Floor(normalized)
		}
		normalizedPrice = PriceType(utils.RoundToDecimalPlace(rounded*math.Pow(10, float64(1-exp)), exp))
	}
	return
}

func NewNormalizedItem(price PriceType, degree int, exp int, roundUp bool, quantityIn ...QuantityType) *NormalizedItem {
	var quantity QuantityType
	if len(quantityIn) == 0 {
		quantity = 0
	} else {
		quantity = quantityIn[0]
	}
	normalizedPrice, err := getNormalizedPrice(price, exp, roundUp)
	if err != nil {
		return nil
	}
	item := &NormalizedItem{
		// Службові дані
		exp:     exp,
		roundUp: roundUp,
		// Дані по ціні
		price:    normalizedPrice,
		quantity: quantity,
		minMax:   btree.New(degree),
		depths:   btree.New(degree)}

	item.SetDepth(NewDepthItem(price, quantity))
	item.SetMinMax(NewQuantityItem(price, quantity, degree))
	return item
}
