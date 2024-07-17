package depth

import (
	"errors"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

func (d *Depth) GetNormalizedAsk(price types.PriceType) (item *types.NormalizedItem, err error) {
	d.Lock()
	defer d.Unlock()
	if val := d.askNormalized.Get(d.NewAskNormalizedItem(price)); val != nil {
		item = val.(*types.NormalizedItem)
	}
	return
}

func (d *Depth) GetNormalizedAskSumma(price types.PriceType) (summa types.QuantityType) {
	d.Lock()
	defer d.Unlock()
	if d.askNormalized != nil {
		askN, _ := d.GetNormalizedAsk(price)
		if askN == nil {
			return
		}
		askN.GetDepths().Ascend(func(i btree.Item) bool {
			summa += i.(*types.DepthItem).GetQuantity()
			return true
		})
	}
	return
}

func (d *Depth) GetNormalizedAsks() *btree.BTree {
	d.Lock()
	defer d.Unlock()
	return d.askNormalized
}

func (d *Depth) GetNormalizedAsksSummaAll() (summa types.QuantityType) {
	d.Lock()
	defer d.Unlock()
	if d.askNormalized != nil {
		d.askNormalized.Ascend(func(i btree.Item) bool {
			summa += i.(*types.NormalizedItem).GetQuantity()
			return true
		})
	}
	return
}

func (d *Depth) GetNormalizedBid(price types.PriceType) (item *types.NormalizedItem, err error) {
	d.Lock()
	defer d.Unlock()
	if val := d.bidNormalized.Get(d.NewBidNormalizedItem(price)); val != nil {
		item = val.(*types.NormalizedItem)
	}
	return
}

func (d *Depth) GetNormalizedBids() *btree.BTree {
	d.Lock()
	defer d.Unlock()
	return d.bidNormalized
}

func (d *Depth) GetNormalizedBidSumma(price types.PriceType) (summa types.QuantityType) {
	d.Lock()
	defer d.Unlock()
	if d.bidNormalized != nil {
		bidN, _ := d.GetNormalizedBid(price)
		if bidN == nil {
			return
		}
		bidN.GetDepths().Ascend(func(i btree.Item) bool {
			summa += i.(*types.NormalizedItem).GetQuantity()
			return true
		})
	}
	return
}

func (d *Depth) GetNormalizedBidsSummaAll() (summa types.QuantityType) {
	d.Lock()
	defer d.Unlock()
	if d.bidNormalized != nil {
		d.bidNormalized.Ascend(func(i btree.Item) bool {
			summa += i.(*types.NormalizedItem).GetQuantity()
			return true
		})
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
	d.Lock()
	defer d.Unlock()
	return d.addNormalized(d.askNormalized, price, quantity, true)
}

func (d *Depth) AddBidNormalized(price types.PriceType, quantity types.QuantityType) error {
	d.Lock()
	defer d.Unlock()
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
	d.Lock()
	defer d.Unlock()
	err = d.deleteNormalized(d.askNormalized, price, quantity, true)
	if err != nil {
		return
	}
	if old := d.askNormalized.Get(d.NewAskNormalizedItem(price)); old != nil {
		if old.(*types.NormalizedItem).IsShouldDelete() {
			d.askNormalized.Delete(d.NewAskNormalizedItem(price))
		}
	}
	return
}

func (d *Depth) DeleteBidNormalized(price types.PriceType, quantity types.QuantityType) (err error) {
	d.Lock()
	defer d.Unlock()
	err = d.deleteNormalized(d.bidNormalized, price, quantity, false)
	if err != nil {
		return
	}
	if old := d.askNormalized.Get(d.NewBidNormalizedItem(price)); old != nil {
		if old.(*types.NormalizedItem).IsShouldDelete() {
			d.bidNormalized.Delete(d.NewBidNormalizedItem(price))
		}
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
