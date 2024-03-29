package depth_analyzer

import (
	"errors"
	"sync"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	types "github.com/fr0ster/go-trading-utils/types"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	DepthAnalyzer struct {
		ask    *btree.BTree
		bid    *btree.BTree
		mu     *sync.Mutex
		Degree int
		Round  int
		Bound  float64
	}
)

func (a *DepthAnalyzer) Lock() {
	a.mu.Lock()
}

func (a *DepthAnalyzer) Unlock() {
	a.mu.Unlock()
}

// Get implements Analyzers.
func (a *DepthAnalyzer) Get(price float64) btree.Item {
	ask := a.ask.Get(&depth_types.DepthItemType{Price: price})
	if ask != nil {
		return ask
	} else {
		return a.bid.Get(&depth_types.DepthItemType{Price: price})
	}
}

// Set implements Analyzers.
func (a *DepthAnalyzer) Set(side types.DepthSide, value btree.Item) {
	if side == types.DepthSideAsk {
		a.ask.ReplaceOrInsert(value)
	} else {
		a.bid.ReplaceOrInsert(value)
	}
}

// AskAscend implements analyzer.DepthAnalyzer.
func (a *DepthAnalyzer) AskAscend(f func(a btree.Item) bool) *btree.BTree {
	a.ask.Ascend(func(a btree.Item) bool {
		return f(a)
	})
	return a.ask
}

// AskDescend implements analyzer.DepthAnalyzer.
func (a *DepthAnalyzer) AskDescend(f func(item btree.Item) bool) *btree.BTree {
	a.ask.Descend(func(a btree.Item) bool {
		return f(a)
	})
	return a.ask
}

// BidAscend implements analyzer.DepthAnalyzer.
func (a *DepthAnalyzer) BidAscend(f func(item btree.Item) bool) *btree.BTree {
	a.bid.Ascend(func(a btree.Item) bool {
		return f(a)
	})
	return a.bid
}

// BidDescend implements analyzer.DepthAnalyzer.
func (a *DepthAnalyzer) BidDescend(f func(item btree.Item) bool) *btree.BTree {
	a.bid.Descend(func(a btree.Item) bool {
		return f(a)
	})
	return a.bid
}

// GetAsks implements analyzer.DepthAnalyzer.
func (a *DepthAnalyzer) GetAsks() *btree.BTree {
	return a.ask
}

// GetBids implements analyzer.DepthAnalyzer.
func (a *DepthAnalyzer) GetBids() *btree.BTree {
	return a.bid
}

// Update implements Analyzers.
func (da *DepthAnalyzer) Update(dp depth_interface.Depth) (err error) {
	if dp == nil {
		return errors.New("DepthAnalyzerLoad returned an empty map")
	}
	da.Lock()
	defer da.Unlock()
	dp.Lock()
	defer dp.Unlock()
	da.bid.Clear(false)
	newBids := btree.New(da.Degree)
	dp.BidDescend(func(item btree.Item) bool {
		bid, _ := Binance2DepthLevels(item)
		bid.Price = utils.RoundToDecimalPlace(bid.Price, da.Round)
		old := newBids.Get(&depth_types.DepthItemType{Price: bid.Price})
		if old != nil {
			bid.Quantity += old.(*depth_types.DepthItemType).Quantity
		}
		newBids.ReplaceOrInsert(bid)
		return true
	})
	newBids.Ascend(func(item btree.Item) bool {
		bid, _ := Binance2DepthLevels(item)
		if bid.Quantity > da.Bound {
			da.bid.ReplaceOrInsert(bid)
		}
		return true
	})
	da.ask.Clear(false)
	newAsks := btree.New(da.Degree)
	dp.AskDescend(func(item btree.Item) bool {
		ask, _ := Binance2DepthLevels(item)
		ask.Price = utils.RoundToDecimalPlace(ask.Price, da.Round)
		old := newAsks.Get(&depth_types.DepthItemType{Price: ask.Price})
		if old != nil {
			ask.Quantity += old.(*depth_types.DepthItemType).Quantity
		}
		newAsks.ReplaceOrInsert(ask)
		return true
	})
	newAsks.Ascend(func(item btree.Item) bool {
		ask, _ := Binance2DepthLevels(item)
		if ask.Quantity > da.Bound {
			da.ask.ReplaceOrInsert(item)
		}
		return true
	})
	return nil
}

func (a *DepthAnalyzer) GetLevels(side types.DepthSide) *btree.BTree {
	getQuantity := func(a btree.Item) float64 {
		if a == nil {
			return 0
		}
		return a.(*depth_types.DepthItemType).Quantity
	}
	ascend := func(dataIn *btree.BTree) (res *btree.BTree) {
		res = btree.New(a.Degree)
		var prev, current, next btree.Item
		dataIn.Ascend(func(a btree.Item) bool {
			next = a
			if (current != nil && prev != nil && getQuantity(current) > getQuantity(prev) && getQuantity(current) > getQuantity(next)) ||
				(current != nil && prev == nil && getQuantity(current) > getQuantity(next)) {
				res.ReplaceOrInsert(current)
			}
			prev = current
			current = next
			return true
		})
		return
	}
	if side == types.DepthSideAsk {
		return ascend(a.ask)
	} else {
		return ascend(a.bid)
	}
}

func NewDepthAnalyzer(degree, round int, bound float64) *DepthAnalyzer {
	return &DepthAnalyzer{
		ask:    btree.New(degree),
		bid:    btree.New(degree),
		mu:     &sync.Mutex{},
		Degree: degree,
		Round:  round,
		Bound:  bound,
	}
}

func Binance2DepthLevels(binanceDepth interface{}) (*depth_types.DepthItemType, error) {
	var val depth_types.DepthItemType
	err := copier.Copy(&val, binanceDepth)
	if err != nil {
		return nil, err
	}
	return &val, nil
}
