package trade_types

import (
	"errors"
	"sync"

	"github.com/google/btree"
)

type (
	AggTrade struct {
		AggTradeID       int64  `json:"a"`
		Price            string `json:"p"`
		Quantity         string `json:"q"`
		FirstTradeID     int64  `json:"f"`
		LastTradeID      int64  `json:"l"`
		Timestamp        int64  `json:"T"`
		IsBuyerMaker     bool   `json:"m"`
		IsBestPriceMatch bool   `json:"M"`
	}
)

func (i *AggTrade) Less(than btree.Item) bool {
	return i.AggTradeID < than.(*AggTrade).AggTradeID
}

func (i *AggTrade) Equal(than btree.Item) bool {
	return i.AggTradeID == than.(*AggTrade).AggTradeID
}

type (
	AggTrades struct {
		tree *btree.BTree
		mu   *sync.Mutex
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

func NewAggTrades() *AggTrades {
	return &AggTrades{
		tree: btree.New(2),
		mu:   &sync.Mutex{},
	}
}

func Binance2AggTrades(binanceTrades interface{}) (*AggTrade, error) {
	switch binanceTrades := binanceTrades.(type) {
	case *AggTrade:
		return binanceTrades, nil
	}
	return nil, errors.New("it's not a AggTrade")
}
