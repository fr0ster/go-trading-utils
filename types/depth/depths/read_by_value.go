package depths

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depths) GetMaxAndSummaByValue(targetSumma items_types.ValueType, up UpOrDown, firstMax ...bool) (
	item *items_types.DepthItem,
	summaValue items_types.ValueType,
	summaQuantity items_types.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		target items_types.ValueType,
		item *items_types.DepthItem,
		summaValue *items_types.ValueType,
		summaQuantity *items_types.QuantityType) func(i btree.Item) bool {
		buffer := items_types.ValueType(0.0)
		return func(i btree.Item) bool {
			if (*summaValue + i.(*items_types.DepthItem).GetValue()) <= target {
				buffer += i.(*items_types.DepthItem).GetValue()
				if !IsFirstMax || i.(*items_types.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*items_types.DepthItem).GetPrice())
					item.SetQuantity(i.(*items_types.DepthItem).GetQuantity())
					*summaQuantity += i.(*items_types.DepthItem).GetQuantity()
					*summaValue = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &items_types.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(targetSumma, item, &summaValue, &summaQuantity))
	} else {
		d.GetTree().Descend(getIterator(targetSumma, item, &summaValue, &summaQuantity))
	}
	return
}

func (d *Depths) GetMaxAndSummaByValuePercent(target float64, up UpOrDown, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	item, value, quantity = d.GetMaxAndSummaByValue(items_types.ValueType(float64(d.GetSummaValue())*target/100), up, firstMax...)
	if value == 0 {
		if up {
			if val := d.GetTree().Min(); val != nil {
				return d.GetMaxAndSummaByPrice(val.(*items_types.DepthItem).GetPrice()*items_types.PriceType(1+target/100), up, firstMax...)
			}
		} else {
			if val := d.GetTree().Max(); val != nil {
				return d.GetMaxAndSummaByPrice(val.(*items_types.DepthItem).GetPrice()*items_types.PriceType(1-target/100), up, firstMax...)
			}
		}
	}
	return
}

func (d *Depths) GetMinMaxByValue(up UpOrDown) (min, max *items_types.DepthItem) {
	getIterator := func(min, max *items_types.DepthItem) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if i.(*items_types.DepthItem).GetValue() >= max.GetValue() {
				max.SetPrice(i.(*items_types.DepthItem).GetPrice())
				max.SetQuantity(i.(*items_types.DepthItem).GetQuantity())
			}
			if i.(*items_types.DepthItem).GetValue() < min.GetValue() || min.GetValue() == 0 {
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
