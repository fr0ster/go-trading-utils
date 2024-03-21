package kline

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	kline_interface "github.com/fr0ster/go-trading-utils/interfaces/kline"
	"github.com/google/btree"
)

type (
	Kline struct {
		btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Kline {
	return &Kline{
		BTree:  *btree.New(int(degree)),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

// Init implements depth_interface.Depths.
func (d *Kline) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) {
	binance.UseTestnet = UseTestnet
	klines, _ :=
		binance.NewClient(apt_key, secret_key).
			NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		d.BTree.ReplaceOrInsert(&kline_interface.Kline{
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
func (d *Kline) GetItem(openTime int64) *kline_interface.Kline {
	return d.BTree.Get(&kline_interface.Kline{OpenTime: int64(openTime)}).(*kline_interface.Kline)
}

// SetItem implements depth_interface.Depths.
func (d *Kline) SetItem(value kline_interface.Kline) {
	d.BTree.ReplaceOrInsert(&value)
}

// Show implements depth_interface.Depths.
func (d *Kline) Show() {
	d.Ascend(func(a btree.Item) bool {
		kline := a.(*kline_interface.Kline)
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
