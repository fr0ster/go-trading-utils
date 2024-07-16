package depth

import (
	"errors"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
)

func (d *Depth) AddAskMinMax(price types.PriceType, quantity types.QuantityType) {
	if d.asksMinMax != nil {
		depthItem := types.NewDepthItem(price, quantity)
		if old := d.asksMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
			old.(*types.QuantityItem).SetDepth(depthItem)
		} else {
			item := d.NewQuantityItem(quantity, price)
			item.SetDepth(depthItem)
			d.asksMinMax.ReplaceOrInsert(item)
		}
	}
}

func (d *Depth) DeleteAskMinMax(price types.PriceType, quantity types.QuantityType) {
	if d.asksMinMax != nil {
		depthItem := types.NewDepthItem(price, quantity)
		if old := d.asksMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
			old.(*types.QuantityItem).DeleteDepth(depthItem)
			d.asksMinMax.Delete(old)
		}
	}
}

func (d *Depth) AddBidMinMax(price types.PriceType, quantity types.QuantityType) {
	if d.bidsMinMax != nil {
		depthItem := types.NewDepthItem(price, quantity)
		if old := d.bidsMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
			old.(*types.QuantityItem).SetDepth(depthItem)
		} else {
			item := d.NewQuantityItem(quantity, price)
			item.SetDepth(depthItem)
			d.bidsMinMax.ReplaceOrInsert(item)
		}
	}
}

func (d *Depth) DeleteBidMinMax(price types.PriceType, quantity types.QuantityType) {
	if d.bidsMinMax != nil {
		depthItem := types.NewDepthItem(price, quantity)
		if old := d.bidsMinMax.Get(d.NewQuantityItem(quantity)); old != nil {
			old.(*types.QuantityItem).DeleteDepth(depthItem)
			d.bidsMinMax.Delete(old)
		}
	}
}

func (d *Depth) AskMin() (min *types.DepthItem, err error) {
	if d.asksMinMax == nil {
		err = errors.New("asksMinMax is nil")
		return
	}
	quantity := d.asksMinMax.Min().(*types.QuantityItem)
	min = quantity.GetDepthMin()
	return
}

func (d *Depth) AskMax() (max *types.DepthItem, err error) {
	if d.asksMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.asksMinMax.Max().(*types.QuantityItem)
	max = quantity.GetDepthMin()
	return
}

func (d *Depth) BidMin() (min *types.DepthItem, err error) {
	if d.bidsMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.bidsMinMax.Min().(*types.QuantityItem)
	min = quantity.GetDepthMax()
	return
}

func (d *Depth) BidMax() (max *types.DepthItem, err error) {
	if d.bidsMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.bidsMinMax.Max().(*types.QuantityItem)
	max = quantity.GetDepthMax()
	return
}

func (d *Depth) NewQuantityItem(quantity types.QuantityType, price ...types.PriceType) *types.QuantityItem {
	if len(price) > 0 {
		return types.NewQuantityItem(price[0], quantity, d.degree)
	} else {
		return types.NewQuantityItem(0, quantity, d.degree)
	}
}
