package depth

import (
	"sync"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	// DepthAnalyzerItem struct {
	// 	Price float64
	// 	// Side     string
	// 	Quantity float64
	// }
	Depth struct {
		asks   *btree.BTree
		bids   *btree.BTree
		mu     *sync.Mutex
		degree int
	}
)

// // DepthItemType - тип для зберігання заявок в стакані
// func (i DepthAnalyzerItem) Less(than btree.Item) bool {
// 	return i.Price < than.(DepthAnalyzerItem).Price
// }

// func (i DepthAnalyzerItem) Equal(than btree.Item) bool {
// 	return i.Price == than.(DepthAnalyzerItem).Price
// }

// func (i *DepthAnalyzerItem) Parse(a common.PriceLevel) {
// 	i.Price, i.Quantity, _ = a.Parse()
// }

func New(degree int) *Depth {
	return &Depth{
		asks:   btree.New(degree),
		bids:   btree.New(degree),
		mu:     &sync.Mutex{},
		degree: degree,
	}
}

func (d *Depth) Lock() {
	d.mu.Lock()
}

func (d *Depth) Unlock() {
	d.mu.Unlock()
}

func (d *Depth) Update(depth depth_interface.Depth) {
	d.Lock()
	defer d.Unlock()
	d.asks.Clear(false)
	d.bids.Clear(false)
	depth.AskAscend(func(a btree.Item) bool {
		d.asks.ReplaceOrInsert(a)
		return true
	})
	depth.BidAscend(func(a btree.Item) bool {
		d.bids.ReplaceOrInsert(a)
		return true
	})
}

// GetBidLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetLevels() *btree.BTree {
	res := btree.New(d.degree)
	getQuantity := func(a btree.Item) float64 {
		if a == nil {
			return 0
		}
		return a.(types.DepthItemType).Quantity
	}
	ascend := func(dataIn, dataOut *btree.BTree) (res *btree.BTree) {
		var prev, current, next btree.Item
		dataIn.Ascend(func(a btree.Item) bool {
			next = a
			if (current != nil && prev != nil && getQuantity(current) > getQuantity(prev) && getQuantity(current) > getQuantity(next)) ||
				(current != nil && prev == nil && getQuantity(current) > getQuantity(next)) {
				dataOut.ReplaceOrInsert(current)
			}
			prev = current
			current = next
			return true
		})
		return
	}
	ascend(d.asks, res)
	ascend(d.bids, res)
	return res
}
