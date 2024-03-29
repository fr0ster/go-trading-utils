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
	"github.com/sirupsen/logrus"
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
	dp.BidDescend(func(item btree.Item) bool {
		bid, _ := Binance2DepthLevels(item)
		bid.Price = utils.RoundToDecimalPlace(bid.Price, da.Round)
		old := da.bid.Get(&depth_types.DepthItemType{Price: bid.Price})
		if old != nil {
			bid.Quantity += old.(*depth_types.DepthItemType).Quantity
		}
		da.bid.ReplaceOrInsert(bid)
		return true
	})
	da.bid.Ascend(func(item btree.Item) bool {
		logrus.Error("item", item, "da.bid.Len()", da.bid.Len())
		if da.bid.Len() > 1 {
			bid, _ := Binance2DepthLevels(item)
			if bid.Quantity < da.Bound {
				da.bid.Delete(item)
			}
			return true
		} else {
			return false
		}
	})
	da.ask.Clear(false)
	dp.AskDescend(func(item btree.Item) bool {
		ask, _ := Binance2DepthLevels(item)
		ask.Price = utils.RoundToDecimalPlace(ask.Price, da.Round)
		old := da.ask.Get(&depth_types.DepthItemType{Price: ask.Price})
		if old != nil {
			ask.Quantity += old.(*depth_types.DepthItemType).Quantity
		}
		da.ask.ReplaceOrInsert(ask)
		return true
	})

	da.ask.Ascend(func(item btree.Item) bool {
		if da.ask.Len() > 1 {
			logrus.Error("item", item, "da.ask.Len()", da.bid.Len())
			ask, _ := Binance2DepthLevels(item)
			if ask.Quantity < da.Bound {
				da.ask.Delete(item)
			}
			return true
		} else {
			return false
		}
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
	// switch val := binanceDepth.(type) {
	// case *depth_types.DepthItemType:
	// 	return val, nil
	// case depth_types.DepthItemType:
	// 	return &val, nil
	// }
	// return nil, errors.New("it's not a DepthLevels")
	var val depth_types.DepthItemType
	err := copier.Copy(&val, binanceDepth)
	if err != nil {
		return nil, err
	}
	return &val, nil
}
