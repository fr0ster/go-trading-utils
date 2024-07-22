package trade

import (
	"sync"
	"time"

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
		startTradeStream func() (chan struct{}, chan struct{}, error)
		Init             func() error
	}
)

// Ascend implements Trades.
func (a *Trades) Ascend(iter func(btree.Item) bool) {
	a.tree.Ascend(iter)
}

// Descend implements Trades.
func (a *Trades) Descend(iter func(btree.Item) bool) {
	a.tree.Descend(iter)
}

// Get implements Trades.
func (a *Trades) Get(id int64) btree.Item {
	res := a.tree.Get(&Trade{ID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements Trades.
func (a *Trades) Lock() {
	a.mu.Lock()
}

// Set implements Trades.
func (a *Trades) Set(val btree.Item) {
	a.tree.ReplaceOrInsert(val)
}

// Unlock implements Trades.
func (a *Trades) Unlock() {
	a.mu.Unlock()
}

// Update implements Trades.
func (a *Trades) Update(val btree.Item) {
	id := val.(*Trade).ID
	old := a.Get(id)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(&Trade{
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

func (a *Trades) GetSymbolname() string {
	return a.symbolname
}

func New(
	stop chan struct{},
	symbolname string,
	startTradeStream func(a *Trades) func() (chan struct{}, chan struct{}, error),
	initCreator func(a *Trades) func() error,
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
