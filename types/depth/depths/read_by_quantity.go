package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depths) GetMaxAndSummaByQuantity(targetSumma types.QuantityType, up UpOrDown, firstMax ...bool) (item *types.DepthItem, summa types.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target types.QuantityType, item *types.DepthItem, summa *types.QuantityType) func(i btree.Item) bool {
		buffer := types.QuantityType(0.0)
		return func(i btree.Item) bool {
			if (*summa + i.(*types.DepthItem).GetQuantity()) < target {
				buffer += i.(*types.DepthItem).GetQuantity()
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

func (d *Depths) GetMaxAndSummaByQuantityPercent(target float64, up UpOrDown, firstMax ...bool) (item *types.DepthItem, summa types.QuantityType) {
	item, summa = d.GetMaxAndSummaByQuantity(types.QuantityType(float64(d.GetSummaQuantity())*target/100), up, firstMax...)
	if summa == 0 {
		if up {
			if val := d.GetTree().Min(); val != nil {
				return d.GetMaxAndSummaByPrice(val.(*types.DepthItem).GetPrice()*types.PriceType(1+target/100), up, firstMax...)
			}
		} else {
			if val := d.GetTree().Max(); val != nil {
				return d.GetMaxAndSummaByPrice(val.(*types.DepthItem).GetPrice()*types.PriceType(1-target/100), up, firstMax...)
			}
		}
	}
	return
}

func (d *Depths) GetMinMaxQuantity(up UpOrDown) (min, max *types.DepthItem) {
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
