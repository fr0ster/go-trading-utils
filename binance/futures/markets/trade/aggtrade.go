package trade

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
)

type (
	AggTradeItem futures.AggTrade
	AggTrades    struct {
		tree btree.BTree
		mu   sync.Mutex
	}
)

func (i *AggTradeItem) Less(than btree.Item) bool {
	return i.AggTradeID < than.(*AggTradeItem).AggTradeID
}

func (i *AggTradeItem) Equal(than btree.Item) bool {
	return i.AggTradeID == than.(*AggTradeItem).AggTradeID
}

// Ascend implements trades.Trades.
func (a *AggTrades) Ascend(iter func(btree.Item) bool) {
	a.tree.Ascend(iter)
}

// Descend implements trades.Trades.
func (a *AggTrades) Descend(iter func(btree.Item) bool) {
	a.tree.Descend(iter)
}

// Get implements trades.Trades.
func (a *AggTrades) Get(val btree.Item) btree.Item {
	res := a.tree.Get(val)
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
	old := a.Get(val)
	if old != nil {
		a.Set(
			&AggTradeItem{
				AggTradeID:   val.(*AggTradeItem).AggTradeID,
				Price:        val.(*AggTradeItem).Price,
				Quantity:     old.(*AggTradeItem).Quantity + val.(*AggTradeItem).Quantity,
				FirstTradeID: val.(*AggTradeItem).FirstTradeID,
				LastTradeID:  val.(*AggTradeItem).LastTradeID,
				Timestamp:    val.(*AggTradeItem).Timestamp,
				IsBuyerMaker: val.(*AggTradeItem).IsBuyerMaker,
			})
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
		new_val := &AggTradeItem{
			AggTradeID:   trade.AggTradeID,
			Price:        trade.Price,
			Quantity:     trade.Quantity,
			FirstTradeID: trade.FirstTradeID,
			LastTradeID:  trade.LastTradeID,
			Timestamp:    trade.Timestamp,
			IsBuyerMaker: trade.IsBuyerMaker}
		a.Update(new_val)
	}
	return nil
}

func NewAggTrades() *AggTrades {
	return &AggTrades{
		tree: *btree.New(2),
		mu:   sync.Mutex{},
	}
}
