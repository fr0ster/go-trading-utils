package tradeV3

import (
	"sync"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
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
		startTradeStream types.StreamFunction
		Init             types.InitFunction
	}
)

// Ascend implements Trades.
func (tv3 *TradesV3) Ascend(iter func(btree.Item) bool) {
	tv3.tree.Ascend(iter)
}

// Descend implements Trades.
func (tv3 *TradesV3) Descend(iter func(btree.Item) bool) {
	tv3.tree.Descend(iter)
}

// Get implements Trades.
func (tv3 *TradesV3) Get(id int64) btree.Item {
	res := tv3.tree.Get(&TradeV3{ID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements Trades.
func (tv3 *TradesV3) Lock() {
	tv3.mu.Lock()
}

// Set implements Trades.
func (tv3 *TradesV3) Set(val btree.Item) {
	tv3.tree.ReplaceOrInsert(val)
}

// Unlock implements Trades.
func (tv3 *TradesV3) Unlock() {
	tv3.mu.Unlock()
}

// Update implements Trades.
func (tv3 *TradesV3) Update(val btree.Item) {
	id := val.(*TradeV3).ID
	old := tv3.Get(id)
	if old == nil {
		tv3.Set(val)
	} else {
		tv3.Set(&TradeV3{
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

func (tv3 *TradesV3) GetSymbolname() string {
	return tv3.symbolname
}

func (tv3 *TradesV3) ResetEvent(err error) {
	tv3.resetEvent <- err
}

func New(
	stop chan struct{},
	symbolname string,
	startTradeStream func(*TradesV3) types.StreamFunction,
	initCreator func(*TradesV3) types.InitFunction) *TradesV3 {
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
