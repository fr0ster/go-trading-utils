package trade

import (
	"sync"

	"github.com/fr0ster/go-trading-utils/interfaces/trades"
	"github.com/google/btree"
	// aggtrade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
)

type (
	// AggTradeItem struct {
	// 	AggTradeID       int64  `json:"a"`
	// 	Price            string `json:"p"`
	// 	Quantity         string `json:"q"`
	// 	FirstTradeID     int64  `json:"f"`
	// 	LastTradeID      int64  `json:"l"`
	// 	Timestamp        int64  `json:"T"`
	// 	IsBuyerMaker     bool   `json:"m"`
	// 	IsBestPriceMatch bool   `json:"M"`
	// }
	AggTrades struct {
		tree btree.BTree
		mu   sync.Mutex
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
func (a *AggTrades) Get(val *trades.AggTradeItem) *trades.AggTradeItem {
	res := a.tree.Get(val)
	if res == nil {
		return nil
	}
	return res.(*trades.AggTradeItem)
}

// // Init implements trades.Trades.
// func (a *AggTrades) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) (err error) {
// 	panic("unimplemented")
// }

// Lock implements trades.Trades.
func (a *AggTrades) Lock() {
	a.mu.Lock()
}

// Set implements trades.Trades.
func (a *AggTrades) Set(val *trades.AggTradeItem) {
	a.tree.ReplaceOrInsert(val)
}

// Unlock implements trades.Trades.
func (a *AggTrades) Unlock() {
	a.mu.Unlock()
}

// Update implements trades.Trades.
func (a *AggTrades) Update(val *trades.AggTradeItem) {
	old := a.Get(
		&trades.AggTradeItem{
			AggTradeID:       val.AggTradeID,
			Price:            val.Price,
			FirstTradeID:     val.FirstTradeID,
			LastTradeID:      val.LastTradeID,
			Quantity:         val.Quantity,
			Timestamp:        val.Timestamp,
			IsBuyerMaker:     val.IsBuyerMaker,
			IsBestPriceMatch: val.IsBestPriceMatch})
	if old == nil {
		a.Set(val)
	} else {
		a.Set(
			&trades.AggTradeItem{
				AggTradeID:       val.AggTradeID,
				Price:            val.Price,
				Quantity:         old.Quantity + val.Quantity,
				FirstTradeID:     val.FirstTradeID,
				LastTradeID:      val.LastTradeID,
				Timestamp:        val.Timestamp,
				IsBuyerMaker:     val.IsBuyerMaker,
				IsBestPriceMatch: val.IsBestPriceMatch})
	}
}

func NewAggTrades() *AggTrades {
	return &AggTrades{
		tree: *btree.New(2),
		mu:   sync.Mutex{},
	}
}
