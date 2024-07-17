package depth

import (
	"errors"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

func (d *Depth) GetNormalizedAsk(price types.PriceType) (item *types.NormalizedItem, err error) {
	normalizedPrice := d.NewAskNormalizedItem(price).GetNormalizedPrice()
	if val := d.askNormalized.Get(d.NewAskNormalizedItem(normalizedPrice)); val != nil {
		item = val.(*types.NormalizedItem)
	}
	return
}

func (d *Depth) GetNormalizedBid(price types.PriceType) (item *types.NormalizedItem, err error) {
	if val := d.bidNormalized.Get(d.NewBidNormalizedItem(price)); val != nil {
		item = val.(*types.NormalizedItem)
	}
	return
}

func (d *Depth) addNormalized(tree *btree.BTree, price types.PriceType, quantity types.QuantityType, RoundUp bool) (err error) {
	if tree != nil {
		if old := tree.Get(d.newNormalizedItem(price, RoundUp)); old != nil {
			// MinMax && Depths
			old.(*types.NormalizedItem).Add(price, quantity)
		} else {
			item := d.newNormalizedItem(price, RoundUp, quantity)
			// item.Add(price, quantity)
			tree.ReplaceOrInsert(item)
		}
	} else {
		err = errors.New("tree is nil")
	}
	return
}

func (d *Depth) AddAskNormalized(price types.PriceType, quantity types.QuantityType) error {
	return d.addNormalized(d.askNormalized, price, quantity, true)
}

func (d *Depth) AddBidNormalized(price types.PriceType, quantity types.QuantityType) error {
	return d.addNormalized(d.bidNormalized, price, quantity, false)
}

func (d *Depth) deleteNormalized(tree *btree.BTree, price types.PriceType, quantity types.QuantityType, roundUp bool) (err error) {
	if tree != nil {
		if old := tree.Get(d.newNormalizedItem(price, roundUp)); old != nil {
			old.(*types.NormalizedItem).Delete(price, quantity)
		}
	} else {
		err = errors.New("tree is nil")
	}
	return
}

func (d *Depth) DeleteAskNormalized(price types.PriceType, quantity types.QuantityType) (err error) {
	err = d.deleteNormalized(d.askNormalized, price, quantity, true)
	if err != nil {
		return
	}
	if d.askNormalized.Get(d.NewAskNormalizedItem(price)).(*types.NormalizedItem).IsEmpty() {
		d.askNormalized.Delete(d.NewAskNormalizedItem(price))
	}
	return
}

func (d *Depth) DeleteBidNormalized(price types.PriceType, quantity types.QuantityType) (err error) {
	err = d.deleteNormalized(d.bidNormalized, price, quantity, false)
	if err != nil {
		return
	}
	if d.askNormalized.Get(d.NewBidNormalizedItem(price)).(*types.NormalizedItem).IsEmpty() {
		d.bidNormalized.Delete(d.NewBidNormalizedItem(price))
	}
	return
}

func (d *Depth) newNormalizedItem(price types.PriceType, roundUp bool, quantity ...types.QuantityType) (normalized *types.NormalizedItem) {
	if len(quantity) > 0 {
		normalized = types.NewNormalizedItem(price, d.degree, d.expBase, roundUp, quantity[0])
	} else {
		normalized = types.NewNormalizedItem(price, d.degree, d.expBase, roundUp)
	}
	return
}

func (d *Depth) NewAskNormalizedItem(price types.PriceType, quantity ...types.QuantityType) (normalized *types.NormalizedItem) {
	normalized = d.newNormalizedItem(price, true, quantity...)
	return
}

func (d *Depth) NewBidNormalizedItem(price types.PriceType, quantity ...types.QuantityType) (normalized *types.NormalizedItem) {
	normalized = d.newNormalizedItem(price, false, quantity...)
	return
}
