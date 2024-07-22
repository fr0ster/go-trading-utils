package tradeV3

import (
	"sync"
	"time"

	"github.com/google/btree"
)

type (
	TradesV3 struct {
		symbolname       string
		tree             *btree.BTree
		mu               *sync.Mutex
		timeOut          time.Duration
		stop             chan struct{}
		resetEvent       chan error
		startTradeStream func() (chan struct{}, chan struct{}, error)
		Init             func() error
	}
)

// Ascend implements Trades.
func (a *TradesV3) Ascend(iter func(btree.Item) bool) {
	a.tree.Ascend(iter)
}

// Descend implements Trades.
func (a *TradesV3) Descend(iter func(btree.Item) bool) {
	a.tree.Descend(iter)
}

// Get implements Trades.
func (a *TradesV3) Get(id int64) btree.Item {
	res := a.tree.Get(&TradeV3{ID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements Trades.
func (a *TradesV3) Lock() {
	a.mu.Lock()
}

// Set implements Trades.
func (a *TradesV3) Set(val btree.Item) {
	a.tree.ReplaceOrInsert(val)
}

// Unlock implements Trades.
func (a *TradesV3) Unlock() {
	a.mu.Unlock()
}

// Update implements Trades.
func (a *TradesV3) Update(val btree.Item) {
	id := val.(*TradeV3).ID
	old := a.Get(id)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(&TradeV3{
			ID:              val.(*TradeV3).ID,
			Symbol:          val.(*TradeV3).Symbol,
			OrderID:         val.(*TradeV3).OrderID,
			OrderListId:     val.(*TradeV3).OrderListId,
			Price:           val.(*TradeV3).Price,
			Quantity:        val.(*TradeV3).Quantity,
			QuoteQuantity:   val.(*TradeV3).QuoteQuantity,
			Commission:      val.(*TradeV3).Commission,
			CommissionAsset: val.(*TradeV3).CommissionAsset,
			Time:            val.(*TradeV3).Time,
			IsBuyer:         val.(*TradeV3).IsBuyer,
			IsMaker:         val.(*TradeV3).IsMaker,
			IsBestMatch:     val.(*TradeV3).IsBestMatch,
			IsIsolated:      val.(*TradeV3).IsIsolated,
		})
	}
}

func (a *TradesV3) GetSymbolname() string {
	return a.symbolname
}

func New(
	stop chan struct{},
	symbolname string,
	startTradeStream func(*TradesV3) func() (chan struct{}, chan struct{}, error),
	initCreator func(*TradesV3) func() error) *TradesV3 {
	this := &TradesV3{
		symbolname: symbolname,
		tree:       btree.New(2),
		mu:         &sync.Mutex{},
		stop:       stop,
	}
	if startTradeStream != nil {
		this.startTradeStream = startTradeStream(this)
	}
	if initCreator != nil {
		this.Init = initCreator(this)
		this.Init()
	}
	return this
}
