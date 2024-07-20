package depths

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depths) GetSummaByQuantity(targetSumma items_types.QuantityType, up UpOrDown, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target items_types.QuantityType, item *items_types.DepthItem, value *items_types.ValueType, quantity *items_types.QuantityType) func(i btree.Item) bool {
		buffer := items_types.QuantityType(0.0)
		return func(i btree.Item) bool {
			if (*quantity + i.(*items_types.DepthItem).GetQuantity()) <= target {
				buffer += i.(*items_types.DepthItem).GetQuantity()
				if !IsFirstMax || i.(*items_types.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*items_types.DepthItem).GetPrice())
					item.SetQuantity(i.(*items_types.DepthItem).GetQuantity())
					*value += i.(*items_types.DepthItem).GetValue()
					*quantity = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &items_types.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(targetSumma, item, &value, &quantity))
	} else {
		d.GetTree().Descend(getIterator(targetSumma, item, &value, &quantity))
	}
	return
}

func (d *Depths) GetSummaByQuantityPercent(target items_types.PricePercentType, up UpOrDown, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	item, value, quantity = d.GetSummaByQuantity(items_types.QuantityType(float64(d.GetSummaQuantity())*float64(target)/100), up, firstMax...)
	if quantity == 0 {
		if up {
			if val := d.GetTree().Min(); val != nil {
				return d.GetSummaByPrice(val.(*items_types.DepthItem).GetPrice()*items_types.PriceType(1+target/100), up, firstMax...)
			}
		} else {
			if val := d.GetTree().Max(); val != nil {
				return d.GetSummaByPrice(val.(*items_types.DepthItem).GetPrice()*items_types.PriceType(1-target/100), up, firstMax...)
			}
		}
	}
	return
}

func (d *Depths) GetMinMaxByQuantity(up UpOrDown) (min, max *items_types.DepthItem) {
	getIterator := func(min, max *items_types.DepthItem) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if i.(*items_types.DepthItem).GetQuantity() >= max.GetQuantity() {
				max.SetPrice(i.(*items_types.DepthItem).GetPrice())
				max.SetQuantity(i.(*items_types.DepthItem).GetQuantity())
			}
			if i.(*items_types.DepthItem).GetQuantity() < min.GetQuantity() || min.GetQuantity() == 0 {
				min.SetPrice(i.(*items_types.DepthItem).GetPrice())
				min.SetQuantity(i.(*items_types.DepthItem).GetQuantity())
			}
			return true
		}
	}
	max = &items_types.DepthItem{}
	min = &items_types.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(min, max))
	} else {
		d.GetTree().Descend(getIterator(min, max))
	}
	return
}
