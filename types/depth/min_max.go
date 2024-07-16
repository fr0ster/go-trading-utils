package depth

import (
	"errors"

	"github.com/google/btree"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
)

func (d *Depth) AddAskMinMax(price float64, quantity float64) {
	if d.asksMinMax != nil {
		depthItem := &types.DepthItem{Price: price, Quantity: quantity}
		if old := d.asksMinMax.Get(&types.QuantityItem{Quantity: quantity}); old != nil {
			old.(*types.QuantityItem).SetDepth(depthItem)
		} else {
			item := &types.QuantityItem{Quantity: quantity, Depths: btree.New(d.degree)}
			item.SetDepth(depthItem)
			d.asksMinMax.ReplaceOrInsert(item)
		}
	}
}

func (d *Depth) DeleteAskMinMax(price float64, quantity float64) {
	if d.asksMinMax != nil {
		depthItem := &types.DepthItem{Price: price, Quantity: quantity}
		if old := d.asksMinMax.Get(&types.QuantityItem{Quantity: quantity}); old != nil {
			old.(*types.QuantityItem).DeleteDepth(depthItem)
			if old.(*types.QuantityItem).Depths.Len() == 0 {
				d.asksMinMax.Delete(old)
			}
		}
	}
}

func (d *Depth) AddBidMinMax(price float64, quantity float64) {
	if d.bidsMinMax != nil {
		depthItem := &types.DepthItem{Price: price, Quantity: quantity}
		if old := d.bidsMinMax.Get(&types.QuantityItem{Quantity: quantity}); old != nil {
			old.(*types.QuantityItem).SetDepth(depthItem)
		} else {
			item := &types.QuantityItem{Quantity: quantity, Depths: btree.New(d.degree)}
			item.SetDepth(depthItem)
			d.bidsMinMax.ReplaceOrInsert(item)
		}
	}
}

func (d *Depth) DeleteBidMinMax(price float64, quantity float64) {
	if d.bidsMinMax != nil {
		depthItem := &types.DepthItem{Price: price, Quantity: quantity}
		if old := d.bidsMinMax.Get(&types.QuantityItem{Quantity: quantity}); old != nil {
			old.(*types.QuantityItem).DeleteDepth(depthItem)
			if old.(*types.QuantityItem).Depths.Len() == 0 {
				d.bidsMinMax.Delete(old)
			}
		}
	}
}

func (d *Depth) AskMin() (min *types.DepthItem, err error) {
	if d.asksMinMax == nil {
		err = errors.New("asksMinMax is nil")
		return
	}
	quantity := d.asksMinMax.Min().(*types.QuantityItem)
	min = quantity.Depths.Min().(*types.DepthItem)
	return
}

func (d *Depth) AskMax() (max *types.DepthItem, err error) {
	if d.asksMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.asksMinMax.Max().(*types.QuantityItem)
	max = quantity.Depths.Min().(*types.DepthItem)
	return
}

func (d *Depth) BidMin() (min *types.DepthItem, err error) {
	if d.bidsMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.bidsMinMax.Min().(*types.QuantityItem)
	min = quantity.Depths.Max().(*types.DepthItem)
	return
}

func (d *Depth) BidMax() (max *types.DepthItem, err error) {
	if d.bidsMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.bidsMinMax.Max().(*types.QuantityItem)
	max = quantity.Depths.Max().(*types.DepthItem)
	return
}
