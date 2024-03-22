package kline

import (
	"sync"

	"github.com/google/btree"
)

type (
	KlineItem struct {
		OpenTime                 int64  `json:"openTime"`
		Open                     string `json:"open"`
		High                     string `json:"high"`
		Low                      string `json:"low"`
		Close                    string `json:"close"`
		Volume                   string `json:"volume"`
		CloseTime                int64  `json:"closeTime"`
		QuoteAssetVolume         string `json:"quoteAssetVolume"`
		TradeNum                 int64  `json:"tradeNum"`
		TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
		TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
	}
	// Kline struct {
	// 	OpenTime                 int64  `json:"openTime"`
	// 	Open                     string `json:"open"`
	// 	High                     string `json:"high"`
	// 	Low                      string `json:"low"`
	// 	Close                    string `json:"close"`
	// 	Volume                   string `json:"volume"`
	// 	CloseTime                int64  `json:"closeTime"`
	// 	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	// 	TradeNum                 int64  `json:"tradeNum"`
	// 	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	// 	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
	// }
	Kline struct {
		tree   btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Kline - тип для зберігання свічок
func (i KlineItem) Less(than btree.Item) bool {
	return i.OpenTime < than.(KlineItem).OpenTime
}

func (i KlineItem) Equal(than btree.Item) bool {
	return i.OpenTime == than.(KlineItem).OpenTime
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
	return d.tree.Get(KlineItem{OpenTime: int64(openTime)})
}

// SetItem implements depth_interface.Depths.
func (d *Kline) Set(value btree.Item) {
	d.tree.ReplaceOrInsert(value)
}

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Kline {
	return &Kline{
		tree:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}
