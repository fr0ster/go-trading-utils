package depth

import (
	"github.com/google/btree"
)

// Відбираємо по ціні
func (d *Depth) GetAsksMaxAndSummaUp(price ...float64) (limit *DepthItem, summa float64) {
	limit = &DepthItem{}
	d.GetAsks().Ascend(func(i btree.Item) bool {
		if len(price) > 0 && i.(*DepthItem).Price <= price[0] {
			summa += i.(*DepthItem).Quantity
			if limit.Quantity < i.(*DepthItem).Quantity {
				limit.Quantity = i.(*DepthItem).Quantity
				limit.Price = i.(*DepthItem).Price
			}
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetBidsMaxAndSummaDown(price ...float64) (limit *DepthItem, summa float64) {
	limit = &DepthItem{}
	d.GetBids().Descend(func(i btree.Item) bool {
		if len(price) > 0 && i.(*DepthItem).Price >= price[0] {
			summa += i.(*DepthItem).Quantity
			if limit.Quantity < i.(*DepthItem).Quantity {
				limit.Quantity = i.(*DepthItem).Quantity
				limit.Price = i.(*DepthItem).Price
			}
			return true
		} else {
			return false
		}
	})
	return
}
