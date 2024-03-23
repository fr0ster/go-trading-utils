package trade_types

import (
	"sync"

	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Trade struct {
		ID            int64  `json:"id"`
		Price         string `json:"price"`
		Quantity      string `json:"qty"`
		QuoteQuantity string `json:"quoteQty"`
		Time          int64  `json:"time"`
		IsBuyerMaker  bool   `json:"isBuyerMaker"`
		IsBestMatch   bool   `json:"isBestMatch"`
		IsIsolated    bool   `json:"isIsolated"`
	}
)

func (i *Trade) Less(than btree.Item) bool {
	return i.ID < than.(*Trade).ID
}

func (i *Trade) Equal(than btree.Item) bool {
	return i.ID == than.(*Trade).ID
}

type (
	Trades struct {
		tree *btree.BTree
		mu   *sync.Mutex
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

func NewTrades() *Trades {
	return &Trades{
		tree: btree.New(2),
		mu:   &sync.Mutex{},
	}
}

func Binance2Trades(binanceTrades interface{}) (*Trade, error) {
	var trade Trade
	err := copier.Copy(&trade, binanceTrades)
	if err != nil {
		return nil, err
	}
	return &trade, nil
}
