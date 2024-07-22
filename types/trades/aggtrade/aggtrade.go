package aggtrade

import (
	"sync"
	"time"

	"github.com/google/btree"
)

type (
	AggTrades struct {
		symbol           string
		tree             *btree.BTree
		mu               *sync.Mutex
		timeOut          time.Duration
		stop             chan struct{}
		resetEvent       chan error
		startTradeStream func() (chan struct{}, chan struct{}, error)
		Init             func() error
	}
)

// Ascend implements trades.Trades.
func (a *AggTrades) Ascend(iter func(btree.Item) bool) {
	a.tree.Ascend(iter)
}

// Descend implements trades.Trades.
func (a *AggTrades) Descend(iter func(btree.Item) bool) {
	a.tree.Descend(iter)
}

// Get implements trades.Trades.
func (a *AggTrades) Get(id int64) btree.Item {
	res := a.tree.Get(&AggTrade{AggTradeID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements trades.Trades.
func (a *AggTrades) Lock() {
	a.mu.Lock()
}

// Set implements trades.Trades.
func (a *AggTrades) Set(val btree.Item) {
	a.tree.ReplaceOrInsert(val)
}

// Unlock implements trades.Trades.
func (a *AggTrades) Unlock() {
	a.mu.Unlock()
}

// Update implements trades.Trades.
func (a *AggTrades) Update(val btree.Item) {
	id := val.(*AggTrade).AggTradeID
	old := a.Get(id)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(&AggTrade{
			AggTradeID:       id,
			Price:            val.(*AggTrade).Price,
			Quantity:         val.(*AggTrade).Quantity,
			FirstTradeID:     val.(*AggTrade).FirstTradeID,
			LastTradeID:      val.(*AggTrade).LastTradeID,
			Timestamp:        val.(*AggTrade).Timestamp,
			IsBuyerMaker:     val.(*AggTrade).IsBuyerMaker,
			IsBestPriceMatch: val.(*AggTrade).IsBestPriceMatch,
		})
	}
}

func (a *AggTrades) Delete(id int64) {
	a.tree.Delete(&AggTrade{AggTradeID: id})
}

func (a *AggTrades) Len() int {
	return a.tree.Len()
}

func (a *AggTrades) Symbol() string {
	return a.symbol
}

func New(
	stop chan struct{},
	symbol string,
	startTradeStream func(*AggTrades) func() (chan struct{}, chan struct{}, error),
	initCreator func(*AggTrades) func() error) *AggTrades {
	this := &AggTrades{
		symbol: symbol,
		tree:   btree.New(2),
		mu:     &sync.Mutex{},
		stop:   stop,
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
