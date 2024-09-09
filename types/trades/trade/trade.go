package trade

import (
	"sync"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	Trades struct {
		symbolname       string
		tree             *btree.BTree
		mu               *sync.Mutex
		timeOut          time.Duration
		stop             chan struct{}
		resetEvent       chan error
		isStartedStream  bool
		startTradeStream types.StreamFunction
		Init             types.InitFunction
	}
)

// Ascend implements Trades.
func (t *Trades) Ascend(iter func(btree.Item) bool) {
	t.tree.Ascend(iter)
}

// Descend implements Trades.
func (t *Trades) Descend(iter func(btree.Item) bool) {
	t.tree.Descend(iter)
}

// Get implements Trades.
func (t *Trades) Get(id int64) btree.Item {
	res := t.tree.Get(&Trade{ID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements Trades.
func (t *Trades) Lock() {
	t.mu.Lock()
}

// Set implements Trades.
func (t *Trades) Set(val btree.Item) {
	t.tree.ReplaceOrInsert(val)
}

// Unlock implements Trades.
func (t *Trades) Unlock() {
	t.mu.Unlock()
}

// Update implements Trades.
func (t *Trades) Update(val btree.Item) {
	id := val.(*Trade).ID
	old := t.Get(id)
	if old == nil {
		t.Set(val)
	} else {
		t.Set(&Trade{
			ID:            id,
			Price:         val.(*Trade).Price,
			Quantity:      val.(*Trade).Quantity,
			QuoteQuantity: val.(*Trade).QuoteQuantity,
			Time:          val.(*Trade).Time,
			IsBuyerMaker:  val.(*Trade).IsBuyerMaker,
			IsBestMatch:   val.(*Trade).IsBestMatch,
			IsIsolated:    val.(*Trade).IsIsolated,
		})
	}
}

func (t *Trades) GetSymbolname() string {
	return t.symbolname
}

func (t *Trades) ResetEvent(err error) {
	t.resetEvent <- err
}

func New(
	stop chan struct{},
	symbolname string,
	startTradeStream func(a *Trades) types.StreamFunction,
	initCreator func(a *Trades) types.InitFunction,
) *Trades {
	this := &Trades{
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
