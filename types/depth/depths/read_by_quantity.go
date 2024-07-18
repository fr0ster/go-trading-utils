package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depths) GetMaxAndSummaByQuantity(targetSumma types.QuantityType, up bool, firstMax ...bool) (item *types.DepthItem, summa types.QuantityType) {
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
