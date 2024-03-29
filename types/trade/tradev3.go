package trade_types

import (
	"errors"
	"sync"

	"github.com/google/btree"
)

type (
	TradeV3 struct {
		ID              int64  `json:"id"`
		Symbol          string `json:"symbol"`
		OrderID         int64  `json:"orderId"`
		OrderListId     int64  `json:"orderListId"`
		Price           string `json:"price"`
		Quantity        string `json:"qty"`
		QuoteQuantity   string `json:"quoteQty"`
		Commission      string `json:"commission"`
		CommissionAsset string `json:"commissionAsset"`
		Time            int64  `json:"time"`
		IsBuyer         bool   `json:"isBuyer"`
		IsMaker         bool   `json:"isMaker"`
		IsBestMatch     bool   `json:"isBestMatch"`
		IsIsolated      bool   `json:"isIsolated"`
	}
)

func (i *TradeV3) Less(than btree.Item) bool {
	return i.ID < than.(*TradeV3).ID
}

func (i *TradeV3) Equal(than btree.Item) bool {
	return i.ID == than.(*TradeV3).ID
}

type (
	TradesV3 struct {
		tree *btree.BTree
		mu   *sync.Mutex
	}
)

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
	res := a.tree.Get(&TradeV3{ID: id})
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
	id := val.(*TradeV3).ID
	old := a.Get(id)
	if old == nil {
		a.Set(val)
	} else {
		a.Set(&TradeV3{
			ID:              val.(*TradeV3).ID,
			Symbol:          val.(*TradeV3).Symbol,
			OrderID:         val.(*TradeV3).OrderID,
			OrderListId:     val.(*TradeV3).OrderListId,
			Price:           val.(*TradeV3).Price,
			Quantity:        val.(*TradeV3).Quantity,
			QuoteQuantity:   val.(*TradeV3).QuoteQuantity,
			Commission:      val.(*TradeV3).Commission,
			CommissionAsset: val.(*TradeV3).CommissionAsset,
			Time:            val.(*TradeV3).Time,
			IsBuyer:         val.(*TradeV3).IsBuyer,
			IsMaker:         val.(*TradeV3).IsMaker,
			IsBestMatch:     val.(*TradeV3).IsBestMatch,
			IsIsolated:      val.(*TradeV3).IsIsolated,
		})
	}
}

func NewTradesV3() *TradesV3 {
	return &TradesV3{
		tree: btree.New(2),
		mu:   &sync.Mutex{},
	}
}

func Binance2TradesV3(binanceTrades interface{}) (*TradeV3, error) {
	switch binanceTrades := binanceTrades.(type) {
	case *TradeV3:
		return binanceTrades, nil
	}
	return nil, errors.New("it's not a TradeV3")
}
