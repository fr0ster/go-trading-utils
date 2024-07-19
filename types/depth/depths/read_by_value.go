package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Depths) GetMaxAndSummaValueByPrice(targetPrice types.PriceType, up UpOrDown, firstMax ...bool) (item *types.DepthItem, summa types.ValueType) {
	var (
		IsFirstMax      bool
		ascendOrDescend func(iterator btree.ItemIterator)
		test            func(price types.PriceType, target types.PriceType) bool
	)
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		targetPrice types.PriceType,
		item *types.DepthItem,
		summa *types.ValueType,
		f func(types.PriceType, types.PriceType) bool) func(i btree.Item) bool {
		buffer := types.ValueType(0.0)
		return func(i btree.Item) bool {
			if f(i.(*types.DepthItem).GetPrice(), targetPrice) {
				buffer += i.(*types.DepthItem).GetValue()
				if !IsFirstMax || types.QuantityType(i.(*types.DepthItem).GetValue()) >= item.GetQuantity() {
					item.SetPrice(i.(*types.DepthItem).GetPrice())
					item.SetQuantity(i.(*types.DepthItem).GetQuantity())
					*summa = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &types.DepthItem{}
	if up {
		ascendOrDescend = d.GetTree().Ascend
		test = func(price types.PriceType, target types.PriceType) bool { return price <= target }
	} else {
		ascendOrDescend = d.GetTree().Descend
		test = func(price types.PriceType, target types.PriceType) bool { return price >= target }
	}
	ascendOrDescend(
		getIterator(targetPrice, item, &summa, test))
	return
}

// Відбираємо по сумі
func (d *Depths) GetMaxAndSummaValue(targetSumma types.ValueType, up UpOrDown, firstMax ...bool) (item *types.DepthItem, summa types.ValueType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target types.ValueType, item *types.DepthItem, summa *types.ValueType) func(i btree.Item) bool {
		buffer := types.ValueType(0.0)
		return func(i btree.Item) bool {
			if (*summa + i.(*types.DepthItem).GetValue()) < target {
				buffer += i.(*types.DepthItem).GetValue()
				if !IsFirstMax || i.(*types.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*types.DepthItem).GetPrice())
					item.SetQuantity(i.(*types.DepthItem).GetQuantity())
					*summa = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	item = &types.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(targetSumma, item, &summa))
	} else {
		d.GetTree().Descend(getIterator(targetSumma, item, &summa))
	}
	return
}

func (d *Depths) GetMaxAndSummaByValuePercent(target float64, up UpOrDown, firstMax ...bool) (item *types.DepthItem, summa types.ValueType) {
	item, summa = d.GetMaxAndSummaValue(types.ValueType(float64(d.GetSummaQuantity())*target/100), up, firstMax...)
	if summa == 0 {
		if up {
			if val := d.GetTree().Min(); val != nil {
				return d.GetMaxAndSummaValueByPrice(val.(*types.DepthItem).GetPrice()*types.PriceType(1+target/100), up, firstMax...)
			}
		} else {
			if val := d.GetTree().Max(); val != nil {
				return d.GetMaxAndSummaValueByPrice(val.(*types.DepthItem).GetPrice()*types.PriceType(1-target/100), up, firstMax...)
			}
		}
	}
	return
}

func (d *Depths) GetMinMaxValue(up UpOrDown) (min, max *types.DepthItem) {
	getIterator := func(min, max *types.DepthItem) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if i.(*types.DepthItem).GetQuantity() >= max.GetQuantity() {
				max.SetPrice(i.(*types.DepthItem).GetPrice())
				max.SetQuantity(i.(*types.DepthItem).GetQuantity())
			}
			if i.(*types.DepthItem).GetQuantity() < min.GetQuantity() || min.GetQuantity() == 0 {
				min.SetPrice(i.(*types.DepthItem).GetPrice())
				min.SetQuantity(i.(*types.DepthItem).GetQuantity())
			}
			return true
		}
	}
	max = &types.DepthItem{}
	min = &types.DepthItem{}
	if up {
		d.GetTree().Ascend(getIterator(min, max))
	} else {
		d.GetTree().Descend(getIterator(min, max))
	}
	return
}
