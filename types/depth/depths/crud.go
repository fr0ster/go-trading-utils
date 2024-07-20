package depths

import (
	"fmt"
	"math"

	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Get implements depth_interface.Depths.
func (a *Depths) Get(item *items_types.DepthItem) *items_types.DepthItem {
	if val := a.tree.Get(item); val != nil {
		return val.(*items_types.DepthItem)
	} else {
		return nil
	}
}

// Set implements depth_interface.Depths.
func (d *Depths) Set(item *items_types.DepthItem) (err error) {
	if old := d.tree.Get(item); old != nil {
		d.summaQuantity += item.GetQuantity() - old.(*items_types.DepthItem).GetQuantity()
		d.summaValue += item.GetValue() - old.(*items_types.DepthItem).GetValue()
	} else {
		d.summaQuantity += item.GetQuantity()
		d.summaValue += item.GetValue()
		d.countQuantity++
	}
	d.tree.ReplaceOrInsert(item)
	return
}

// Delete implements depth_interface.Depths.
func (d *Depths) Delete(item *items_types.DepthItem) {
	old := d.tree.Get(item)
	if old != nil {
		d.summaQuantity -= old.(*items_types.DepthItem).GetQuantity()
		d.summaValue -= old.(*items_types.DepthItem).GetValue()
		d.countQuantity--
		d.tree.Delete(item)
	}
}

// Update implements depth_interface.Depths.
func (d *Depths) Update(item *items_types.DepthItem) bool {
	if item.GetQuantity() == 0 {
		d.Delete(item)
	} else {
		d.Set(item)
	}
	return true
}

// Count implements depth_interface.Depths.
func (d *Depths) Count() int {
	return d.countQuantity
}

func (d *Depths) GetSummaQuantity() items_types.QuantityType {
	return d.summaQuantity
}

func (d *Depths) GetSummaValue() items_types.ValueType {
	return d.summaValue
}

func (d *Depths) GetMiddleQuantity() items_types.QuantityType {
	return d.summaQuantity / items_types.QuantityType(d.countQuantity)
}

func (d *Depths) GetMiddleValue() items_types.ValueType {
	return d.summaValue / items_types.ValueType(d.countQuantity)
}

func (d *Depths) GetMinPrice() (min *items_types.DepthItem, err error) {
	if val := d.GetTree().Min(); val != nil {
		min = val.(*items_types.DepthItem)
	} else {
		err = fmt.Errorf("Depths is empty")
	}
	return
}

func (d *Depths) GetMaxPrice() (max *items_types.DepthItem, err error) {
	if val := d.GetTree().Max(); val != nil {
		max = val.(*items_types.DepthItem)
	} else {
		err = fmt.Errorf("Depths is empty")
	}
	return
}

func (d *Depths) GetDeltaPrice() (delta items_types.PriceType, err error) {
	max, err := d.GetMaxPrice()
	if err != nil {
		return
	}
	min, err := d.GetMinPrice()
	if err != nil {
		return
	}
	delta = max.GetPrice() - min.GetPrice()
	return
}

func (d *Depths) GetDeltaPricePercent() (delta items_types.PricePercentType, err error) {
	deltaPrice, err := d.GetDeltaPrice()
	if err != nil {
		return
	}
	min, err := d.GetMinPrice()
	if err != nil {
		return
	}
	return items_types.PricePercentType(deltaPrice/min.GetPrice()) * 100, nil
}

func (d *Depths) GetStandardDeviation() float64 {
	summaSquares := 0.0
	d.GetTree().Ascend(func(i btree.Item) bool {
		depth := i.(*items_types.DepthItem)
		summaSquares += depth.GetQuantityDeviation(d.GetMiddleQuantity()) * depth.GetQuantityDeviation(d.GetMiddleQuantity())
		return true
	})
	return math.Sqrt(summaSquares / float64(d.Count()))
}

func (pp *Depths) NextPriceUp(percent items_types.PricePercentType) items_types.PriceType {
	min, _ := pp.GetMinPrice()
	return min.GetPrice() * (1 + items_types.PriceType(percent)/100)
}

func (pp *Depths) NextPriceDown(percent items_types.PricePercentType) items_types.PriceType {
	max, _ := pp.GetMaxPrice()
	return max.GetPrice() * (1 - items_types.PriceType(percent)/100)
}
