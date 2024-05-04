package kline

import (
	"sync"

	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Kline struct {
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
		IsFinal                  bool   `json:"x"`
	}
	Klines struct {
		Time   int64
		tree   *btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Kline - тип для зберігання свічок
func (i *Kline) Less(than btree.Item) bool {
	return i.OpenTime < than.(*Kline).OpenTime
}

func (i *Kline) Equal(than btree.Item) bool {
	return i.OpenTime == than.(*Kline).OpenTime
}

func (d *Klines) Ascend(f func(btree.Item) bool) {
	d.tree.Ascend(f)
}

func (d *Klines) Descend(f func(btree.Item) bool) {
	d.tree.Descend(f)
}

// Lock implements kline_interface.Klines.
func (d *Klines) Lock() {
	d.mutex.Lock()
}

// Unlock implements kline_interface.Klines.
func (d *Klines) Unlock() {
	d.mutex.Unlock()
}

// GetItem implements kline_interface.Klines.
func (d *Klines) GetKline(openTime int64) btree.Item {
	return d.tree.Get(&Kline{OpenTime: int64(openTime)})
}

// SetItem implements kline_interface.Klines.
func (d *Klines) SetKline(value btree.Item) {
	d.tree.ReplaceOrInsert(value)
}

// GetKlines implements kline_interface.Klines.
func (d *Klines) GetKlines() *btree.BTree {
	return d.tree
}

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Klines {
	return &Klines{
		tree:   btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func Binance2kline(binanceKline interface{}) (*Kline, error) {
	var val Kline
	err := copier.Copy(&val, binanceKline)
	if err != nil {
		return nil, err
	}
	return &val, nil
}
