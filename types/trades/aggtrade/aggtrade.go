package aggtrade

import (
	"sync"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
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
		startTradeStream types.StreamFunction
		Init             types.InitFunction
	}
)

// Ascend implements trades.Trades.
func (at *AggTrades) Ascend(iter func(btree.Item) bool) {
	at.tree.Ascend(iter)
}

// Descend implements trades.Trades.
func (at *AggTrades) Descend(iter func(btree.Item) bool) {
	at.tree.Descend(iter)
}

// Get implements trades.Trades.
func (at *AggTrades) Get(id int64) btree.Item {
	res := at.tree.Get(&AggTrade{AggTradeID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements trades.Trades.
func (at *AggTrades) Lock() {
	at.mu.Lock()
}

// Set implements trades.Trades.
func (at *AggTrades) Set(val btree.Item) {
	at.tree.ReplaceOrInsert(val)
}

// Unlock implements trades.Trades.
func (at *AggTrades) Unlock() {
	at.mu.Unlock()
}

// Update implements trades.Trades.
func (at *AggTrades) Update(val btree.Item) {
	id := val.(*AggTrade).AggTradeID
	old := at.Get(id)
	if old == nil {
		at.Set(val)
	} else {
		at.Set(&AggTrade{
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

func (at *AggTrades) Delete(id int64) {
	at.tree.Delete(&AggTrade{AggTradeID: id})
}

func (a *AggTrades) Len() int {
	return a.tree.Len()
}

func (at *AggTrades) Symbol() string {
	return at.symbol
}

func (at *AggTrades) ResetEvent(err error) {
	at.resetEvent <- err
}

func New(
	stop chan struct{},
	symbol string,
	startTradeStream func(*AggTrades) types.StreamFunction,
	initCreator func(*AggTrades) types.InitFunction) *AggTrades {
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
