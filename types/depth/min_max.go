package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
)

// func (d *Depth) AddAskMinMax(price types.PriceType, quantity types.QuantityType) {
// 	if d.asksMinMax != nil {
// 		if old := d.asksMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
// 			old.(*types.QuantityItem).Add(price, quantity)
// 		} else {
// 			item := d.NewQuantityItem(quantity, price)
// 			item.Add(price, quantity)
// 			d.asksMinMax.ReplaceOrInsert(item)
// 		}
// 	}
// }

// func (d *Depth) DeleteAskMinMax(price types.PriceType, quantity types.QuantityType) {
// 	if d.asksMinMax != nil {
// 		if old := d.asksMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
// 			old.(*types.QuantityItem).Delete(price, quantity)
// 			if old.(*types.QuantityItem).IsShouldDelete() {
// 				d.asksMinMax.Delete(old)
// 			}
// 		}
// 	}
// }

// func (d *Depth) AddBidMinMax(price types.PriceType, quantity types.QuantityType) {
// 	if d.bidsMinMax != nil {
// 		if old := d.bidsMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
// 			old.(*types.QuantityItem).Add(price, quantity)
// 		} else {
// 			item := d.NewQuantityItem(quantity, price)
// 			item.Add(price, quantity)
// 			d.bidsMinMax.ReplaceOrInsert(item)
// 		}
// 	}
// }

// func (d *Depth) DeleteBidMinMax(price types.PriceType, quantity types.QuantityType) {
// 	if d.bidsMinMax != nil {
// 		if old := d.bidsMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
// 			old.(*types.QuantityItem).Delete(price, quantity)
// 			if old.(*types.QuantityItem).IsShouldDelete() {
// 				d.bidsMinMax.Delete(old)
// 			}
// 		}
// 	}
// }

// func (d *Depth) AskMin() (min *types.DepthItem, err error) {
// 	if d.asksMinMax == nil {
// 		err = errors.New("asksMinMax is nil")
// 		return
// 	}
// 	if quantity := d.asksMinMax.Min(); quantity != nil {
// 		min = quantity.(*types.QuantityItem).GetDepthMin()
// 	} else {
// 		err = errors.New("asksMinMax is empty")
// 	}
// 	return
// }

// func (d *Depth) AskMax() (max *types.DepthItem, err error) {
// 	if d.asksMinMax == nil {
// 		err = errors.New("asksMinMax is empty")
// 		return
// 	}
// 	if quantity := d.asksMinMax.Max(); quantity != nil {
// 		max = quantity.(*types.QuantityItem).GetDepthMin()
// 	} else {
// 		err = errors.New("asksMinMax is empty")
// 	}
// 	return
// }

// func (d *Depth) BidMin() (min *types.DepthItem, err error) {
// 	if d.bidsMinMax == nil {
// 		err = errors.New("bidsMinMax is empty")
// 		return
// 	}
// 	if quantity := d.bidsMinMax.Min(); quantity != nil {
// 		min = quantity.(*types.QuantityItem).GetDepthMax()
// 	} else {
// 		err = errors.New("bidsMinMax is empty")
// 	}
// 	return
// }

// func (d *Depth) BidMax() (max *types.DepthItem, err error) {
// 	if d.bidsMinMax == nil {
// 		err = errors.New("bidsMinMax is empty")
// 		return
// 	}
// 	if quantity := d.bidsMinMax.Max(); quantity != nil {
// 		max = quantity.(*types.QuantityItem).GetDepthMax()
// 	} else {
// 		err = errors.New("bidsMinMax is empty")
// 	}
// 	return
// }

func (d *Depth) NewQuantityItem(quantity types.QuantityType, price ...types.PriceType) *types.QuantityItem {
	if len(price) > 0 {
		return types.NewQuantityItem(price[0], quantity, d.degree)
	} else {
		return types.NewQuantityItem(0, quantity, d.degree)
	}
}
