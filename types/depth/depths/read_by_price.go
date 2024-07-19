package depths

import (
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	// types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Depths) GetMaxAndSummaByPrice(targetPrice items.PriceType, up UpOrDown, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	var (
		IsFirstMax      bool
		ascendOrDescend func(iterator btree.ItemIterator)
		test            func(price items.PriceType, target items.PriceType) bool
	)
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		targetPrice items.PriceType,
		item *items.DepthItem,
		value *items.ValueType,
		quantity *items.QuantityType,
		f func(items.PriceType, items.PriceType) bool) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if f(i.(*items.DepthItem).GetPrice(), targetPrice) {
				if !IsFirstMax || i.(*items.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*items.DepthItem).GetPrice())
					item.SetQuantity(i.(*items.DepthItem).GetQuantity())
					*value += i.(*items.DepthItem).GetValue()
					*quantity += i.(*items.DepthItem).GetQuantity()
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &items.DepthItem{}
	if up {
		ascendOrDescend = d.GetTree().Ascend
		test = func(price items.PriceType, target items.PriceType) bool { return price <= target }
	} else {
		ascendOrDescend = d.GetTree().Descend
		test = func(price items.PriceType, target items.PriceType) bool { return price >= target }
	}
	ascendOrDescend(
		getIterator(targetPrice, item, &value, &quantity, test))
	return
}

func (d *Depths) GetMinMaxByPrice(up UpOrDown) (min, max *items.DepthItem) {
	getIterator := func(min, max *items.DepthItem) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if i.(*items.DepthItem).GetPrice() >= max.GetPrice() {
				max.SetPrice(i.(*items.DepthItem).GetPrice())
				max.SetQuantity(i.(*items.DepthItem).GetQuantity())
			}
			if i.(*items.DepthItem).GetPrice() < min.GetPrice() || min.GetPrice() == 0 {
				min.SetPrice(i.(*items.DepthItem).GetPrice())
				min.SetQuantity(i.(*items.DepthItem).GetQuantity())
			}
			return true
		}
	}
	max = &items.DepthItem{}
	min = &items.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(min, max))
	} else {
		d.GetTree().Descend(getIterator(min, max))
	}
	return
}
