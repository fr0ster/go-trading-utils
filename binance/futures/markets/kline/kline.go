package kline

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	// kline_interface "github.com/fr0ster/go-trading-utils/interfaces/kline"
	"github.com/google/btree"
)

type (
	KlineItem futures.Kline
	Kline     struct {
		tree   btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Less implements btree.Item.
func (k KlineItem) Less(than btree.Item) bool {
	return k.OpenTime < than.(*KlineItem).OpenTime
}

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Kline {
	return &Kline{
		tree:   *btree.New(int(degree)),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func (d *Kline) Ascend(iter func(btree.Item) bool) {
	d.tree.Ascend(iter)
}

func (d *Kline) Descend(iter func(btree.Item) bool) {
	d.tree.Descend(iter)
}

func (d *Kline) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) {
	futures.UseTestnet = UseTestnet
	klines, _ :=
		futures.NewClient(apt_key, secret_key).NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		d.tree.ReplaceOrInsert(KlineItem(*kline))
	}
}

// Lock implements depth_interface.Depths.
func (d *Kline) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Kline) Unlock() {
	d.mutex.Unlock()
}

// GetItem implements depth_interface.Depths.
func (d *Kline) Get(openTime int64) btree.Item {
	return d.tree.Get(&KlineItem{OpenTime: int64(openTime)})
}

// SetItem implements depth_interface.Depths.
func (d *Kline) Set(value btree.Item) {
	d.tree.ReplaceOrInsert(value)
}
