package depths

import (
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depths) GetMaxAndSummaByQuantity(targetSumma items.QuantityType, up UpOrDown, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target items.QuantityType, item *items.DepthItem, value *items.ValueType, quantity *items.QuantityType) func(i btree.Item) bool {
		buffer := items.QuantityType(0.0)
		return func(i btree.Item) bool {
			if (*quantity + i.(*items.DepthItem).GetQuantity()) <= target {
				buffer += i.(*items.DepthItem).GetQuantity()
				if !IsFirstMax || i.(*items.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*items.DepthItem).GetPrice())
					item.SetQuantity(i.(*items.DepthItem).GetQuantity())
					*value += i.(*items.DepthItem).GetValue()
					*quantity = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &items.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(targetSumma, item, &value, &quantity))
	} else {
		d.GetTree().Descend(getIterator(targetSumma, item, &value, &quantity))
	}
	return
}

func (d *Depths) GetMaxAndSummaByQuantityPercent(target float64, up UpOrDown, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	item, value, quantity = d.GetMaxAndSummaByQuantity(items.QuantityType(float64(d.GetSummaQuantity())*target/100), up, firstMax...)
	if quantity == 0 {
		if up {
			if val := d.GetTree().Min(); val != nil {
				return d.GetMaxAndSummaByPrice(val.(*items.DepthItem).GetPrice()*items.PriceType(1+target/100), up, firstMax...)
			}
		} else {
			if val := d.GetTree().Max(); val != nil {
				return d.GetMaxAndSummaByPrice(val.(*items.DepthItem).GetPrice()*items.PriceType(1-target/100), up, firstMax...)
			}
		}
	}
	return
}

func (d *Depths) GetMinMaxByQuantity(up UpOrDown) (min, max *items.DepthItem) {
	getIterator := func(min, max *items.DepthItem) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if i.(*items.DepthItem).GetQuantity() >= max.GetQuantity() {
				max.SetPrice(i.(*items.DepthItem).GetPrice())
				max.SetQuantity(i.(*items.DepthItem).GetQuantity())
			}
			if i.(*items.DepthItem).GetQuantity() < min.GetQuantity() || min.GetQuantity() == 0 {
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
