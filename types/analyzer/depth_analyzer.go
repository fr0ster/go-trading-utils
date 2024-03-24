package analyzer

import (
	"sync"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	types "github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	DepthAnalyzer struct {
		TargetPrice    float64
		TargetQuantity float64
		ask            *btree.BTree
		bid            *btree.BTree
		mu             sync.Mutex
		degree         int
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
func (a *DepthAnalyzer) Update(dp depth_interface.Depth) error {
	a.Lock()
	defer a.Unlock()
	dp.BidDescend(func(item btree.Item) bool {
		bid, _ := Binance2DepthLevels(item)
		a.bid.ReplaceOrInsert(bid)
		return true
	})
	dp.AskDescend(func(item btree.Item) bool {
		ask, _ := Binance2DepthLevels(item)
		a.ask.ReplaceOrInsert(ask)
		return true
	})
	return nil
}

func (a *DepthAnalyzer) GetLevels() *btree.BTree {

	res := btree.New(a.degree)
	getQuantity := func(a btree.Item) float64 {
		if a == nil {
			return 0
		}
		return a.(*types.DepthLevels).Quantity
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
	ascend(a.ask, res)
	ascend(a.bid, res)
	return res
}

func NewDepthAnalyzer(degree int) *DepthAnalyzer {
	return &DepthAnalyzer{
		ask:    btree.New(degree),
		bid:    btree.New(degree),
		mu:     sync.Mutex{},
		degree: degree,
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
