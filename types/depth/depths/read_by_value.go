package depths

import (
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depths) GetMaxAndSummaByValue(targetSumma items.ValueType, up UpOrDown, firstMax ...bool) (
	item *items.DepthItem,
	summaValue items.ValueType,
	summaQuantity items.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		target items.ValueType,
		item *items.DepthItem,
		summaValue *items.ValueType,
		summaQuantity *items.QuantityType) func(i btree.Item) bool {
		buffer := items.ValueType(0.0)
		return func(i btree.Item) bool {
			if (*summaValue + i.(*items.DepthItem).GetValue()) <= target {
				buffer += i.(*items.DepthItem).GetValue()
				if !IsFirstMax || i.(*items.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*items.DepthItem).GetPrice())
					item.SetQuantity(i.(*items.DepthItem).GetQuantity())
					*summaQuantity += i.(*items.DepthItem).GetQuantity()
					*summaValue = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &items.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(targetSumma, item, &summaValue, &summaQuantity))
	} else {
		d.GetTree().Descend(getIterator(targetSumma, item, &summaValue, &summaQuantity))
	}
	return
}

func (d *Depths) GetMaxAndSummaByValuePercent(target float64, up UpOrDown, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	item, value, quantity = d.GetMaxAndSummaByValue(items.ValueType(float64(d.GetSummaValue())*target/100), up, firstMax...)
	if value == 0 {
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

func (d *Depths) GetMinMaxByValue(up UpOrDown) (min, max *items.DepthItem) {
	getIterator := func(min, max *items.DepthItem) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if i.(*items.DepthItem).GetValue() >= max.GetValue() {
				max.SetPrice(i.(*items.DepthItem).GetPrice())
				max.SetQuantity(i.(*items.DepthItem).GetQuantity())
			}
			if i.(*items.DepthItem).GetValue() < min.GetValue() || min.GetValue() == 0 {
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
