package trade

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	TradeItem binance.Trade
	Trades    struct {
		tree btree.BTree
		mu   sync.Mutex
	}
)

func (i *TradeItem) Less(than btree.Item) bool {
	return i.ID < than.(*TradeItem).ID
}

func (i *TradeItem) Equal(than btree.Item) bool {
	return i.ID == than.(*TradeItem).ID
}

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
	res := a.tree.Get(&TradeItem{ID: id})
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
	old := a.Get(val.(*TradeItem).ID)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(
			&TradeItem{
				ID:            val.(*TradeItem).ID,
				Price:         val.(*TradeItem).Price,
				Quantity:      old.(*TradeItem).Quantity + val.(*TradeItem).Quantity,
				QuoteQuantity: val.(*TradeItem).QuoteQuantity,
				Time:          val.(*TradeItem).Time,
				IsBuyerMaker:  val.(*TradeItem).IsBuyerMaker,
				IsBestMatch:   val.(*TradeItem).IsBestMatch,
				IsIsolated:    val.(*TradeItem).IsIsolated})

	}
}

func NewTrades() *Trades {
	return &Trades{
		tree: *btree.New(2),
		mu:   sync.Mutex{},
	}
}
func tradesInit(res []*binance.Trade, a *Trades) (err error) {
	for _, val := range res {
		trade := val
		a.Update(&TradeItem{
			ID:            trade.ID,
			Price:         trade.Price,
			Quantity:      trade.Quantity,
			QuoteQuantity: trade.QuoteQuantity,
			Time:          trade.Time,
			IsBuyerMaker:  trade.IsBuyerMaker,
			IsBestMatch:   trade.IsBestMatch,
			IsIsolated:    trade.IsIsolated})
	}
	return nil
}

func HistoricalTradesInit(a *Trades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewHistoricalTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesInit(res, a)
}

func RecentTradesInit(a *Trades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewRecentTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesInit(res, a)
}
