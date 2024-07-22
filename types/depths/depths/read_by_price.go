package depths

import (
	"math"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/google/btree"
)

func (d *Depths) GetSummaByPrice(targetPrice items_types.PriceType, up UpOrDown, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	var (
		IsFirstMax      bool
		ascendOrDescend func(iterator btree.ItemIterator)
		test            func(price items_types.PriceType, target items_types.PriceType) bool
	)
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		targetPrice items_types.PriceType,
		item *items_types.DepthItem,
		value *items_types.ValueType,
		quantity *items_types.QuantityType,
		f func(items_types.PriceType, items_types.PriceType) bool) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if f(i.(*items_types.DepthItem).GetPrice(), targetPrice) {
				if !IsFirstMax || i.(*items_types.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*items_types.DepthItem).GetPrice())
					item.SetQuantity(i.(*items_types.DepthItem).GetQuantity())
					*value += i.(*items_types.DepthItem).GetValue()
					*quantity += i.(*items_types.DepthItem).GetQuantity()
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &items_types.DepthItem{}
	if up {
		ascendOrDescend = d.GetTree().Ascend
		test = func(price items_types.PriceType, target items_types.PriceType) bool { return price <= target }
	} else {
		ascendOrDescend = d.GetTree().Descend
		test = func(price items_types.PriceType, target items_types.PriceType) bool { return price >= target }
	}
	ascendOrDescend(
		getIterator(targetPrice, item, &value, &quantity, test))
	return
}

func (d *Depths) GetSummaByPricePercent(target items_types.PricePercentType, up UpOrDown, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	var price items_types.PriceType
	max, err := d.GetMaxPrice()
	if err != nil {
		return
	}
	min, err := d.GetMinPrice()
	if err != nil {
		return
	}
	delta := items_types.PriceType(math.Abs(float64(max.GetPrice()-min.GetPrice()) * float64(target) / 100))
	if up {
		price = min.GetPrice() + delta
	} else {
		price = max.GetPrice() - delta
	}
	item, value, quantity = d.GetSummaByPrice(price, up, firstMax...)
	return
}

func (d *Depths) GetMinMaxByPrice(up UpOrDown) (min, max *items_types.DepthItem, err error) {
	max, err = d.GetMaxPrice()
	if err != nil {
		return
	}
	min, err = d.GetMinPrice()
	return
}

func (d *Depths) GetSummaByPriceRange(first, last items_types.PriceType, f ...items_types.DepthFilter) (
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	var (
		filter func(*items_types.DepthItem) bool
		ranger func(lessOrEqual, greaterThan btree.Item, iterator btree.ItemIterator)
	)
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*items_types.DepthItem) bool { return true }
	}
	if first <= last {
		ranger = d.GetTree().DescendRange
	} else {
		ranger = d.GetTree().AscendRange
	}
	ranger(items_types.New(last), items_types.New(first), func(i btree.Item) bool {
		if filter(i.(*items_types.DepthItem)) {
			quantity += i.(*items_types.DepthItem).GetQuantity()
			value += i.(*items_types.DepthItem).GetValue()
		}
		return true
	})
	return
}
