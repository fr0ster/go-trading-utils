package depth

import (
	"errors"

	"github.com/google/btree"
)

func (d *Depth) AddAskMinMax(price float64, quantity float64) {
	if d.asksMinMax != nil {
		depthItem := &DepthItem{Price: price, Quantity: quantity}
		if old := d.asksMinMax.Get(&QuantityItem{Quantity: quantity}); old != nil {
			old.(*QuantityItem).Depths.ReplaceOrInsert(depthItem)
		} else {
			item := &QuantityItem{Quantity: quantity, Depths: btree.New(d.degree)}
			item.Depths.ReplaceOrInsert(depthItem)
			d.asksMinMax.ReplaceOrInsert(item)
		}
	}
}

func (d *Depth) DeleteAskMinMax(price float64, quantity float64) {
	if d.asksMinMax != nil {
		depthItem := &DepthItem{Price: price, Quantity: quantity}
		if old := d.asksMinMax.Get(&QuantityItem{Quantity: quantity}); old != nil {
			old.(*QuantityItem).Depths.Delete(depthItem)
			if old.(*QuantityItem).Depths.Len() == 0 {
				d.asksMinMax.Delete(old)
			}
		}
	}
}

func (d *Depth) AddBidMinMax(price float64, quantity float64) {
	if d.bidsMinMax != nil {
		depthItem := &DepthItem{Price: price, Quantity: quantity}
		if old := d.bidsMinMax.Get(&QuantityItem{Quantity: quantity}); old != nil {
			old.(*QuantityItem).Depths.ReplaceOrInsert(depthItem)
		} else {
			item := &QuantityItem{Quantity: quantity, Depths: btree.New(d.degree)}
			item.Depths.ReplaceOrInsert(depthItem)
			d.bidsMinMax.ReplaceOrInsert(item)
		}
	}
}

func (d *Depth) DeleteBidMinMax(price float64, quantity float64) {
	if d.bidsMinMax != nil {
		depthItem := &DepthItem{Price: price, Quantity: quantity}
		if old := d.bidsMinMax.Get(&QuantityItem{Quantity: quantity}); old != nil {
			old.(*QuantityItem).Depths.Delete(depthItem)
			if old.(*QuantityItem).Depths.Len() == 0 {
				d.bidsMinMax.Delete(old)
			}
		}
	}
}

func (d *Depth) AskMin() (min *DepthItem, err error) {
	if d.asksMinMax == nil {
		err = errors.New("asksMinMax is nil")
		return
	}
	quantity := d.asksMinMax.Min().(*QuantityItem)
	min = quantity.Depths.Min().(*DepthItem)
	return
}

func (d *Depth) AskMax() (max *DepthItem, err error) {
	if d.asksMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.asksMinMax.Max().(*QuantityItem)
	max = quantity.Depths.Min().(*DepthItem)
	return
}

func (d *Depth) BidMin() (min *DepthItem, err error) {
	if d.bidsMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.bidsMinMax.Min().(*QuantityItem)
	min = quantity.Depths.Max().(*DepthItem)
	return
}

func (d *Depth) BidMax() (max *DepthItem, err error) {
	if d.bidsMinMax == nil {
		err = errors.New("asksMinMax is empty")
		return
	}
	quantity := d.bidsMinMax.Max().(*QuantityItem)
	max = quantity.Depths.Max().(*DepthItem)
	return
}
