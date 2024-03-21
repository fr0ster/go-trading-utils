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
		client *binance.Client
		tree   btree.BTree
		mu     sync.Mutex
	}
)

func (i *TradeV3Item) Less(than btree.Item) bool {
	return i.ID < than.(*TradeV3Item).ID
}

func (i *TradeV3Item) Equal(than btree.Item) bool {
	return i.ID == than.(*TradeV3Item).ID
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
func (a *TradesV3) Get(val btree.Item) btree.Item {
	res := a.tree.Get(val)
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
	old := a.Get(val)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(
			&TradeV3Item{
				ID:              old.(*TradeV3Item).ID,
				Symbol:          old.(*TradeV3Item).Symbol,
				OrderID:         old.(*TradeV3Item).OrderID,
				OrderListId:     old.(*TradeV3Item).OrderListId,
				Price:           old.(*TradeV3Item).Price,
				Quantity:        old.(*TradeV3Item).Quantity,
				QuoteQuantity:   old.(*TradeV3Item).QuoteQuantity,
				Commission:      old.(*TradeV3Item).Commission,
				CommissionAsset: old.(*TradeV3Item).CommissionAsset,
				Time:            old.(*TradeV3Item).Time,
				IsBuyer:         old.(*TradeV3Item).IsBuyer,
				IsMaker:         old.(*TradeV3Item).IsMaker,
				IsBestMatch:     old.(*TradeV3Item).IsBestMatch,
				IsIsolated:      old.(*TradeV3Item).IsIsolated})

	}
}

func NewTradesV3() *TradesV3 {
	return &TradesV3{
		tree: *btree.New(2),
		mu:   sync.Mutex{},
	}
}

func tradesV3Init(res []*binance.TradeV3, a *TradesV3, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	for _, val := range res {
		old := val
		a.Update(&TradeV3Item{
			ID:              old.ID,
			Symbol:          old.Symbol,
			OrderID:         old.OrderID,
			OrderListId:     old.OrderListId,
			Price:           old.Price,
			Quantity:        old.Quantity,
			QuoteQuantity:   old.QuoteQuantity,
			Commission:      old.Commission,
			CommissionAsset: old.CommissionAsset,
			Time:            old.Time,
			IsBuyer:         old.IsBuyer,
			IsMaker:         old.IsMaker,
			IsBestMatch:     old.IsBestMatch,
			IsIsolated:      old.IsIsolated})
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
	return tradesV3Init(res, a, secret_key, symbolname, limit, UseTestnet)
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
	return tradesV3Init(res, a, secret_key, symbolname, limit, UseTestnet)
}
