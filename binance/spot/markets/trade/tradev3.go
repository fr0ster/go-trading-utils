package trade

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	TradeV3Item binance.TradeV3
	TradesV3    struct {
		tree *btree.BTree
		mu   *sync.Mutex
	}
)

func (i TradeV3Item) Less(than btree.Item) bool {
	return i.ID < than.(TradeV3Item).ID
}

func (i TradeV3Item) Equal(than btree.Item) bool {
	return i.ID == than.(TradeV3Item).ID
}

// Ascend implements Trades.
func (a *TradesV3) Ascend(iter func(btree.Item) bool) {
	a.tree.Ascend(iter)
}

// Descend implements Trades.
func (a *TradesV3) Descend(iter func(btree.Item) bool) {
	a.tree.Descend(iter)
}

// Get implements Trades.
func (a *TradesV3) Get(id int64) btree.Item {
	res := a.tree.Get(TradeV3Item{ID: id})
	if res == nil {
		return nil
	}
	return res
}

// Lock implements Trades.
func (a *TradesV3) Lock() {
	a.mu.Lock()
}

// Set implements Trades.
func (a *TradesV3) Set(val btree.Item) {
	a.tree.ReplaceOrInsert(val)
}

// Unlock implements Trades.
func (a *TradesV3) Unlock() {
	a.mu.Unlock()
}

// Update implements Trades.
func (a *TradesV3) Update(val btree.Item) {
	old := a.Get(val.(TradeV3Item).ID)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(old.(TradeV3Item))
	}
}

func NewTradesV3() *TradesV3 {
	return &TradesV3{
		tree: btree.New(2),
		mu:   &sync.Mutex{},
	}
}

func tradesV3Init(res []*binance.TradeV3, a *TradesV3) (err error) {
	for _, val := range res {
		a.Update(TradeV3Item(*val))
	}
	return nil
}

func ListTradesInit(a *TradesV3, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewListTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesV3Init(res, a)
}

func ListMarginTradesInit(a *TradesV3, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewListMarginTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesV3Init(res, a)
}
