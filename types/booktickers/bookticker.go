package booktickers

import (
	"sync"
	"time"

	items_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"

	"github.com/google/btree"
)

type (
	BookTickers struct {
		symbol                string
		tree                  *btree.BTree
		mutex                 sync.Mutex
		degree                int
		timeOut               time.Duration
		startBookTickerStream func() (chan struct{}, chan struct{}, error)
		init                  func() error
		stop                  chan struct{}
		resetEvent            chan error
	}
)

func (btt *BookTickers) Lock() {
	btt.mutex.Lock()
}

func (btt *BookTickers) Unlock() {
	btt.mutex.Unlock()
}

func (btt *BookTickers) TryLock() bool {
	return btt.mutex.TryLock()
}

func (btt *BookTickers) Ascend(f func(item btree.Item) bool) {
	btt.tree.Ascend(f)
}

func (btt *BookTickers) Descend(f func(item btree.Item) bool) {
	btt.tree.Descend(f)
}

func (btt *BookTickers) Get(symbol string) (item *items_types.BookTicker) {
	if val := btt.tree.Get(&items_types.BookTicker{Symbol: symbol}); val != nil {
		item = val.(*items_types.BookTicker)
	}
	return
}

func (btt *BookTickers) Set(item *items_types.BookTicker) {
	btt.tree.ReplaceOrInsert(item)
}

func (btt *BookTickers) GetSymbol() string {
	return btt.symbol
}

func New(
	stop chan struct{},
	degree int,
	startBookTickerStreamCreator func(*BookTickers) func() (chan struct{}, chan struct{}, error),
	initCreator func(*BookTickers) func() error,
	symbols ...string) *BookTickers {
	var symbol string
	if len(symbols) > 0 {
		symbol = symbols[0]
	}
	this := &BookTickers{
		symbol:     symbol,
		tree:       btree.New(degree),
		mutex:      sync.Mutex{},
		degree:     degree,
		timeOut:    1 * time.Hour,
		stop:       stop,
		resetEvent: make(chan error),
	}
	if startBookTickerStreamCreator != nil {
		this.startBookTickerStream = startBookTickerStreamCreator(this)
	}
	if initCreator != nil {
		this.init = initCreator(this)
		this.init()
	}

	return this
}
