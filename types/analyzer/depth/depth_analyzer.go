package depth_analyzer

import (
	"sync"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	types "github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	DepthAnalyzer struct {
		ask    *btree.BTree
		bid    *btree.BTree
		mu     sync.Mutex
		degree int
		round  int
		bound  float64
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
	ask := a.ask.Get(&types.DepthLevels{Price: price})
	if ask != nil {
		return ask
	} else {
		return a.bid.Get(&types.DepthLevels{Price: price})
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

// Update implements Analyzers.
func (da *DepthAnalyzer) Update(dp depth_interface.Depth) (err error) {
	da.Lock()
	defer da.Unlock()
	da.bid.Clear(false)
	dp.BidDescend(func(item btree.Item) bool {
		bid, _ := Binance2DepthLevels(item)
		bid.Price = utils.RoundToDecimalPlace(bid.Price, da.round)
		old := da.bid.Get(&types.DepthLevels{Price: bid.Price})
		if old != nil {
			bid.Quantity += old.(*types.DepthLevels).Quantity
		}
		da.bid.ReplaceOrInsert(bid)
		return true
	})
	var bid *types.DepthLevels
	da.bid.Ascend(func(item btree.Item) bool {
		bid, err = Binance2DepthLevels(item)
		if bid.Quantity < da.bound {
			da.bid.Delete(item)
		}
		return true
	})
	da.ask.Clear(false)
	dp.AskDescend(func(item btree.Item) bool {
		ask, _ := Binance2DepthLevels(item)
		ask.Price = utils.RoundToDecimalPlace(ask.Price, da.round)
		old := da.ask.Get(&types.DepthLevels{Price: ask.Price})
		if old != nil {
			ask.Quantity += old.(*types.DepthLevels).Quantity
		}
		da.ask.ReplaceOrInsert(ask)
		return true
	})
	var ask *types.DepthLevels
	da.ask.Ascend(func(item btree.Item) bool {
		ask, err = Binance2DepthLevels(item)
		if err != nil {
			return false
		}
		if ask.Quantity < da.bound {
			da.ask.Delete(item)
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
		return a.(*types.DepthLevels).Quantity
	}
	ascend := func(dataIn *btree.BTree) (res *btree.BTree) {
		res = btree.New(a.degree)
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
		mu:     sync.Mutex{},
		degree: degree,
		round:  round,
		bound:  bound,
	}
}

func Binance2DepthLevels(binanceDepth interface{}) (*types.DepthLevels, error) {
	var depthLevelItem types.DepthLevels
	err := copier.Copy(&depthLevelItem, binanceDepth)
	if err != nil {
		return nil, err
	}
	return &depthLevelItem, nil
}
