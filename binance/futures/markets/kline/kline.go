package kline

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	// kline_interface "github.com/fr0ster/go-trading-utils/interfaces/kline"
	"github.com/google/btree"
)

type (
	KlineItem futures.Kline
	Kline     struct {
		btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Less implements btree.Item.
func (k *KlineItem) Less(than btree.Item) bool {
	return k.OpenTime < than.(*KlineItem).OpenTime
}

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Kline {
	return &Kline{
		BTree:  *btree.New(int(degree)),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func (d *Kline) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) {
	futures.UseTestnet = UseTestnet
	klines, _ :=
		futures.NewClient(apt_key, secret_key).NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		d.BTree.ReplaceOrInsert(&KlineItem{
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

// Lock implements depth_interface.Depths.
func (d *Kline) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Kline) Unlock() {
	d.mutex.Unlock()
}

// GetItem implements depth_interface.Depths.
func (d *Kline) GetItem(openTime int64) *KlineItem {
	return d.BTree.Get(&KlineItem{OpenTime: int64(openTime)}).(*KlineItem)
}

// SetItem implements depth_interface.Depths.
func (d *Kline) SetItem(value KlineItem) {
	d.BTree.ReplaceOrInsert(&value)
}

// Show implements depth_interface.Depths.
func (d *Kline) Show() {
	d.Ascend(func(a btree.Item) bool {
		kline := a.(*KlineItem)
		fmt.Printf(
			"OpenTime: %d, Open: %s, High: %s, Low: %s, Close: %s, Volume: %s, CloseTime: %d, QuoteAssetVolume: %s, TradeNum: %d, TakerBuyBaseAssetVolume: %s, TakerBuyQuoteAssetVolume: %s\n",
			kline.OpenTime,
			kline.Open,
			kline.High,
			kline.Low,
			kline.Close,
			kline.Volume,
			kline.CloseTime,
			kline.QuoteAssetVolume,
			kline.TradeNum,
			kline.TakerBuyBaseAssetVolume,
			kline.TakerBuyQuoteAssetVolume,
		)
		return true
	})
}
