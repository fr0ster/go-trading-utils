package kline

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	KlineItem binance.Kline
	Kline     struct {
		tree   btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Kline - тип для зберігання свічок
func (i *KlineItem) Less(than btree.Item) bool {
	return i.OpenTime < than.(*KlineItem).OpenTime
}

func (i *KlineItem) Equal(than btree.Item) bool {
	return i.OpenTime == than.(*KlineItem).OpenTime
}

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Kline {
	return &Kline{
		tree:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func (d *Kline) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) {
	binance.UseTestnet = UseTestnet
	klines, _ :=
		binance.NewClient(apt_key, secret_key).
			NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		d.tree.ReplaceOrInsert(&KlineItem{
			OpenTime:                 kline.OpenTime,
			Open:                     kline.Open,
			High:                     kline.High,
			Low:                      kline.Low,
			Close:                    kline.Close,
			Volume:                   kline.Volume,
			CloseTime:                kline.CloseTime,
			QuoteAssetVolume:         kline.QuoteAssetVolume,
			TradeNum:                 kline.TradeNum,
			TakerBuyBaseAssetVolume:  kline.TakerBuyBaseAssetVolume,
			TakerBuyQuoteAssetVolume: kline.TakerBuyQuoteAssetVolume,
		})
	}
}

func (d *Kline) Ascend(f func(btree.Item) bool) {
	d.tree.Ascend(f)
}

func (d *Kline) Descend(f func(btree.Item) bool) {
	d.tree.Descend(f)
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
