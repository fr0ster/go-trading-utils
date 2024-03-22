package trade

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	// types.AggTrade futures.AggTrade
	AggTrades struct {
		tree btree.BTree
		mu   sync.Mutex
	}
)

// func (i types.AggTrade) Less(than btree.Item) bool {
// 	return i.AggTradeID < than.(*types.AggTrade).AggTradeID
// }

// func (i types.AggTrade) Equal(than btree.Item) bool {
// 	return i.AggTradeID == than.(*types.AggTrade).AggTradeID
// }

// Ascend implements AggTrades.
func (a *AggTrades) Ascend(iter func(btree.Item) bool) {
	a.tree.Ascend(iter)
}

// Descend implements AggTrades.
func (a *AggTrades) Descend(iter func(btree.Item) bool) {
	a.tree.Descend(iter)
}

// Get implements AggTrades.
func (a *AggTrades) Get(id int64) btree.Item {
	res := a.tree.Get(&types.AggTrade{AggTradeID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements AggTrades.
func (a *AggTrades) Lock() {
	a.mu.Lock()
}

// Set implements AggTrades.
func (a *AggTrades) Set(val btree.Item) {
	a.tree.ReplaceOrInsert(val)
}

// Unlock implements AggTrades.
func (a *AggTrades) Unlock() {
	a.mu.Unlock()
}

// Update implements AggTrades.
func (a *AggTrades) Update(val btree.Item) {
	old := a.Get(val.(*types.AggTrade).AggTradeID)
	if old != nil {
		a.Set(val.(*types.AggTrade))
	} else {
		a.Set(val)
	}
}

func AggTradeInit(a *AggTrades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	futures.UseTestnet = UseTestnet
	client := futures.NewClient(apt_key, secret_key)
	res, err :=
		client.NewAggTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	for _, trade := range res {
		a.Update(types.AggTrade{
			AggTradeID:   trade.AggTradeID,
			Price:        trade.Price,
			Quantity:     trade.Quantity,
			Timestamp:    trade.Timestamp,
			IsBuyerMaker: trade.IsBuyerMaker,
		})
	}
	return nil
}

func NewAggTrades() *AggTrades {
	return &AggTrades{
		tree: *btree.New(2),
		mu:   sync.Mutex{},
	}
}
