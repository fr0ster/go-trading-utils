package prices

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
)

type (
	PriceChangeStatsItem futures.PriceChangeStats
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
		futures.NewClient(apt_key, secret_key).
			NewListPriceChangeStatsService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, pcs := range pcss {
		d.tree.ReplaceOrInsert(PriceChangeStatsItem(*pcs))
	}
}

func (d *PriceChangeStats) Lock() {
	d.mutex.Lock()
}

func (d *PriceChangeStats) Unlock() {
	d.mutex.Unlock()
}
