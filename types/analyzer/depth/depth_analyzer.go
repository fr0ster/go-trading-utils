package depth_analyzer

import (
	"errors"
	"sync"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	types "github.com/fr0ster/go-trading-utils/types"
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
		bid.Price = utils.RoundToDecimalPlace(bid.Price, da.round)
		old := da.bid.Get(&types.DepthLevels{Price: bid.Price})
		if old != nil {
			bid.Quantity += old.(*types.DepthLevels).Quantity
		}
		da.bid.ReplaceOrInsert(bid)
		return true
	})
	var bid *types.DepthLevels
	if da.bid.Len() != 0 {
		da.bid.Ascend(func(item btree.Item) bool {
			if item != nil {
				bid, err = Binance2DepthLevels(item)
				if da.bid != nil && bid.Quantity < da.bound {
					da.bid.Delete(item)
				}
			}
			return true
		})
	}
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
	if da.ask.Len() != 0 {
		logrus.Errorf("Ascend begin Ask item: %v", da.ask.Len())
		da.ask.Ascend(func(item btree.Item) bool {
			logrus.Errorf("Ascend begin Ask item: %v", item)
			if item != nil {
				ask, err = Binance2DepthLevels(item)
				if err != nil {
					return false
				}
				if da.ask != nil && ask.Quantity < da.bound {
					da.ask.Delete(item)
				}
			}
			return true
		})
		logrus.Debug("Ascend end Ask item: ", da.ask)
	}
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
		mu:     &sync.Mutex{},
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
