package depth

import (
	"sync"

	"github.com/adshao/go-binance/v2/common"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/google/btree"
)

type (
	DepthItemType struct {
		Price    float64
		Side     string
		Quantity float64
	}
	PriceLevelType struct {
		Price    float64
		Quantity float64
	}
	Depth struct {
		tree   *btree.BTree
		mu     sync.Mutex
		degree int
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(DepthItemType).Price
}

func (i DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(DepthItemType).Price
}

func (i *DepthItemType) Parse(a common.PriceLevel) {
	i.Price, i.Quantity, _ = a.Parse()
}

// PriceLevelType - тип для зберігання заявок в стакані
func (i PriceLevelType) Less(than btree.Item) bool {
	return i.Price < than.(PriceLevelType).Price
}

func (i PriceLevelType) Equal(than btree.Item) bool {
	return i.Price == than.(PriceLevelType).Price
}

func New(degree int) *Depth {
	return &Depth{
		tree:   btree.New(degree),
		mu:     sync.Mutex{},
		degree: degree,
	}
}

func (d *Depth) Lock() {
	d.mu.Lock()
}

func (d *Depth) Unlock() {
	d.mu.Unlock()
}

func (d *Depth) Update(a depth_interface.Depth) {
	d.Lock()
	defer d.Unlock()
	d.tree.Clear(false)
	a.AskAscend(func(a btree.Item) bool {
		pl := a.(PriceLevelType)
		d.tree.ReplaceOrInsert(DepthItemType{Side: "ask", Price: pl.Price, Quantity: pl.Quantity})
		return true
	})
	a.BidAscend(func(a btree.Item) bool {
		pl := a.(PriceLevelType)
		d.tree.ReplaceOrInsert(DepthItemType{Side: "bid", Price: pl.Price, Quantity: pl.Quantity})
		return true
	})
}

// GetBidLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetLevels() *btree.BTree {
	maximaTree := btree.New(d.degree)
	func() {
		var prev, current, next *PriceLevelType
		d.tree.Ascend(func(a btree.Item) bool {
			next = a.(*PriceLevelType)
			if (current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity) ||
				(current != nil && prev == nil && current.Quantity > next.Quantity) {
				maximaTree.ReplaceOrInsert(current)
			}
			prev = current
			current = next
			return true
		})
	}()
	func() {
		var prev, current, next *PriceLevelType
		d.tree.Ascend(func(a btree.Item) bool {
			next = a.(*PriceLevelType)
			if (current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity) ||
				(current != nil && prev == nil && current.Quantity > next.Quantity) {
				maximaTree.ReplaceOrInsert(current)
			}
			prev = current
			current = next
			return true
		})
	}()
	return maximaTree
}
