package depth

import (
	"errors"
	"math"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

func (d *Depth) GetNormalizedPrice(price float64, RoundUp bool) (normalizedPrice float64, err error) {
	getNormalizedPrice := func(price float64, max float64) float64 {
		len := int(math.Log10(max))
		exp := 2
		rounded := 0.0
		if len == exp {
			return price
		} else if len > exp {
			normalized := price * math.Pow(10, float64(-exp))
			if RoundUp {
				rounded = math.Ceil(normalized)
			} else {
				rounded = math.Floor(normalized)
			}
			return rounded * math.Pow(10, float64(exp))
		} else {
			return price * math.Pow(10, float64(exp))
		}
	}
	if max := d.asks.Max(); max != nil {
		normalizedPrice = utils.RoundToDecimalPlace(getNormalizedPrice(price, max.(*types.DepthItem).Price), 0)
	} else if max := d.bids.Max(); max != nil {
		normalizedPrice = utils.RoundToDecimalPlace(getNormalizedPrice(price, max.(*types.DepthItem).Price), 0)
	} else {
		err = errors.New("asks and bids is empty")
	}
	return
}

func (d *Depth) GetNormalizedAsk(price float64) (item *types.NormalizedItem, err error) {
	normalizedPrice, err := d.GetNormalizedPrice(price, false)
	if err != nil {
		return
	}
	if val := d.askNormalized.Get(&types.NormalizedItem{Price: normalizedPrice}); val != nil {
		item = val.(*types.NormalizedItem)
	}
	return
}

func (d *Depth) GetNormalizedBid(price float64) (item *types.NormalizedItem, err error) {
	normalizedPrice, err := d.GetNormalizedPrice(price, true)
	if err != nil {
		return
	}
	if val := d.bidNormalized.Get(&types.NormalizedItem{Price: normalizedPrice}); val != nil {
		item = val.(*types.NormalizedItem)
	}
	return
}

func (d *Depth) addNormalized(tree *btree.BTree, price float64, quantity float64, RoundUp bool) (err error) {
	var normalizedPrice float64
	if tree != nil {
		normalizedPrice, err = d.GetNormalizedPrice(price, RoundUp)
		if err != nil {
			return
		}
		depthItem := &types.DepthItem{Price: price, Quantity: quantity}
		if old := tree.Get(&types.NormalizedItem{Price: normalizedPrice}); old != nil {
			if val := old.(*types.NormalizedItem).GetMinMax(quantity); val != nil {
				val.SetDepth(&types.DepthItem{Price: price, Quantity: quantity})
			} else {
				item := &types.QuantityItem{Quantity: quantity, Depths: btree.New(d.degree)}
				item.SetDepth(depthItem)
				old.(*types.NormalizedItem).SetMinMax(item)
			}
			old.(*types.NormalizedItem).SetDepth(depthItem)
			old.(*types.NormalizedItem).Quantity += quantity
		} else {
			item := &types.NormalizedItem{
				Price:     normalizedPrice,
				Quantity:  quantity,
				MinMax:    btree.New(d.degree),
				DepthItem: btree.New(d.degree),
			}
			minMax := &types.QuantityItem{Quantity: quantity, Depths: btree.New(d.degree)}
			minMax.SetDepth(depthItem)
			item.SetMinMax(minMax)
			item.SetDepth(depthItem)
			tree.ReplaceOrInsert(item)
		}
	} else {
		err = errors.New("tree is nil")
	}
	return
}

func (d *Depth) AddAskNormalized(price float64, quantity float64) error {
	return d.addNormalized(d.askNormalized, price, quantity, false)
}

func (d *Depth) AddBidNormalized(price float64, quantity float64) error {
	return d.addNormalized(d.bidNormalized, price, quantity, true)
}

func (d *Depth) deleteNormalized(tree *btree.BTree, price float64, quantity float64) (err error) {
	if tree != nil {
		depthItem := &types.DepthItem{Price: price, Quantity: quantity}
		if old := tree.Get(&types.NormalizedItem{Price: price}); old != nil {
			if val := old.(*types.NormalizedItem).GetMinMax(quantity); val != nil {
				val.DeleteDepth(depthItem)
				if val.Depths.Len() == 0 {
					old.(*types.NormalizedItem).DeleteMinMax(val)
				}
			}
			old.(*types.NormalizedItem).DeleteDepth(depthItem)
		}
	} else {
		err = errors.New("tree is nil")
	}
	return
}

func (d *Depth) DeleteAskNormalized(price float64, quantity float64) error {
	return d.deleteNormalized(d.askNormalized, price, quantity)
}

func (d *Depth) DeleteBidNormalized(price float64, quantity float64) error {
	return d.deleteNormalized(d.bidNormalized, price, quantity)
}

func (d *Depth) GetNormalizedAsks() *btree.BTree {
	return d.askNormalized
}

func (d *Depth) GetNormalizedBids() *btree.BTree {
	return d.bidNormalized
}
