package prices

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	// prices_interface "github.com/fr0ster/go-trading-utils/interfaces/prices"
	"github.com/google/btree"
)

type (
	PriceChangeStatsItem binance.PriceChangeStats
	PriceChangeStats     struct {
		tree   btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Less implements btree.Item.
func (p PriceChangeStatsItem) Less(than btree.Item) bool {
	return p.OpenTime < than.(*PriceChangeStatsItem).OpenTime
}

func (d *PriceChangeStats) Get(symbol string) btree.Item {
	return d.tree.Get(&PriceChangeStatsItem{Symbol: symbol})
}

func (d *PriceChangeStats) Set(value btree.Item) {
	d.tree.ReplaceOrInsert(value)
}

// PriceChangeStats - B-дерево для зберігання Цінових змін
func New(degree int) *PriceChangeStats {
	return &PriceChangeStats{
		tree:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func (d *PriceChangeStats) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) {
	binance.UseTestnet = UseTestnet
	pcss, _ :=
		binance.NewClient(apt_key, secret_key).
			NewListPriceChangeStatsService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, pcs := range pcss {
		d.tree.ReplaceOrInsert(&PriceChangeStatsItem{
			Symbol:             pcs.Symbol,
			PriceChange:        pcs.PriceChange,
			PriceChangePercent: pcs.PriceChangePercent,
			WeightedAvgPrice:   pcs.WeightedAvgPrice,
			PrevClosePrice:     pcs.PrevClosePrice,
			LastPrice:          pcs.LastPrice,
			LastQty:            pcs.LastQty,
			BidPrice:           pcs.BidPrice,
			BidQty:             pcs.BidQty,
			AskPrice:           pcs.AskPrice,
			AskQty:             pcs.AskQty,
			OpenPrice:          pcs.OpenPrice,
			HighPrice:          pcs.HighPrice,
			LowPrice:           pcs.LowPrice,
			Volume:             pcs.Volume,
			QuoteVolume:        pcs.QuoteVolume,
			OpenTime:           pcs.OpenTime,
			CloseTime:          pcs.CloseTime,
			FirstID:            pcs.FirstID,
			LastID:             pcs.LastID,
			Count:              pcs.Count,
		})

	}
}

func (d *PriceChangeStats) Lock() {
	d.mutex.Lock()
}

func (d *PriceChangeStats) Unlock() {
	d.mutex.Unlock()
}
