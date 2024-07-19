package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Depths) GetMaxAndSummaQuantityByPrice(targetPrice types.PriceType, up UpOrDown, firstMax ...bool) (item *types.DepthItem, quantity types.QuantityType) {
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
		summa *types.QuantityType,
		f func(types.PriceType, types.PriceType) bool) func(i btree.Item) bool {
		buffer := types.QuantityType(0.0)
		return func(i btree.Item) bool {
			if f(i.(*types.DepthItem).GetPrice(), targetPrice) {
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
		ascendOrDescend = d.GetTree().Ascend
		test = func(price types.PriceType, target types.PriceType) bool { return price <= target }
	} else {
		ascendOrDescend = d.GetTree().Descend
		test = func(price types.PriceType, target types.PriceType) bool { return price >= target }
	}
	ascendOrDescend(
		getIterator(targetPrice, item, &quantity, test))
	return
}
