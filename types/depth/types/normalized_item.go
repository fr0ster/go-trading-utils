package types

import (
	"math"

	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	NormalizedItem struct {
		// Службові дані
		degree  int
		exp     int
		roundUp bool
		// Дані по ціні
		price    PriceType
		quantity QuantityType
		minMax   *btree.BTree
		depths   *btree.BTree
	}
)

// Функції для btree.Btree
func (i *NormalizedItem) Less(than btree.Item) bool {
	return i.price < than.(*NormalizedItem).price
}

func (i *NormalizedItem) Equal(than btree.Item) bool {
	return i.price == than.(*NormalizedItem).price
}

// CRUD
func (i *NormalizedItem) GetNormalizedPrice() PriceType {
	return i.price
}

func (i *NormalizedItem) Add(price PriceType, quantity QuantityType) {
	normalizedPrice, err := getNormalizedPrice(price, i.exp, i.roundUp)
	if err == nil && normalizedPrice == i.price {
		i.quantity += quantity
		i.depths.ReplaceOrInsert(NewDepthItem(price, i.quantity))
		i.minMax.ReplaceOrInsert(NewQuantityItem(normalizedPrice, quantity, i.degree))
	}
}

func (i *NormalizedItem) Delete(price PriceType, quantity QuantityType) {
	if old := i.GetDepth(price); old != nil {
		i.quantity -= quantity
	}
	i.minMax.Delete(NewQuantityItem(price, quantity, i.degree))
	i.depths.Delete(NewDepthItem(price))
}

// Робота з кількістю по нормалізованим ордерам, повинно дорівнувати суммі кількостей по всіх ордерах в стакані
func (i *NormalizedItem) GetQuantity() QuantityType {
	return i.quantity
}

func (i *NormalizedItem) SetQuantity(quantity QuantityType) {
	i.quantity = quantity
}

func (i *NormalizedItem) GetMinMaxes() *btree.BTree {
	return i.minMax
}

// Робота зі стаканом
func (i *NormalizedItem) GetDepth(price PriceType) (depthItem *DepthItem) {
	if i.depths != nil {
		depthItem = i.depths.Get(NewDepthItem(price)).(*DepthItem)
	}
	return
}

func (i *NormalizedItem) GetDepths() *btree.BTree {
	return i.depths
}

// Робота з Мінімальними та Максимальними значеннями
func (i *NormalizedItem) GetMinMax(quantity QuantityType) (minMax *QuantityItem) {
	if i.minMax != nil {
		if val := i.minMax.Get(&QuantityItem{quantity: quantity}); val != nil {
			minMax = val.(*QuantityItem)
		}
	}
	return
}

func (i *NormalizedItem) IsShouldDelete() bool {
	return i.depths == nil && i.minMax == nil
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
		degree:  degree,
		exp:     exp,
		roundUp: roundUp,
		// Дані по ціні
		price:    normalizedPrice,
		quantity: quantity,
		minMax:   btree.New(degree),
		depths:   btree.New(degree)}
	if quantity != 0 {
		item.minMax.ReplaceOrInsert(NewQuantityItem(normalizedPrice, quantity, degree))
		item.depths.ReplaceOrInsert(NewDepthItem(price, quantity))
	}
	return item
}
