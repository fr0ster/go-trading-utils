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
		interval     string
		time         int64
		klines_final *btree.BTree
		last_kline   *Kline
		mutex        sync.Mutex
		degree       int
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
	d.klines_final.Ascend(f)
}

func (d *Klines) Descend(f func(btree.Item) bool) {
	d.klines_final.Descend(f)
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
	return d.klines_final.Get(&Kline{OpenTime: int64(openTime)})
}

// SetItem implements kline_interface.Klines.
func (d *Klines) SetKline(value btree.Item) {
	d.klines_final.ReplaceOrInsert(value)
}

// GetLastKline implements kline_interface.Klines.
func (d *Klines) GetLastKline() *Kline {
	return d.last_kline
}

// SetLastKline implements kline_interface.Klines.
func (d *Klines) SetLastKline(value *Kline) {
	d.last_kline = value
}

// GetKlines implements kline_interface.Klines.
func (d *Klines) GetKlines() *btree.BTree {
	return d.klines_final
}

// GetTime implements kline_interface.Klines.
func (d *Klines) GetTime() int64 {
	return d.time
}

// SetTime implements kline_interface.Klines.
func (d *Klines) SetTime(time int64) {
	d.time = time
}

// GetInterval implements kline_interface.Klines.
func (d *Klines) GetInterval() string {
	return d.interval
}

// SetInterval implements kline_interface.Klines.
func (d *Klines) SetInterval(interval string) {
	d.interval = interval
}

// Kline - B-дерево для зберігання стакана заявок
func New(degree int, interval string) *Klines {
	return &Klines{
		interval:     interval,
		time:         0,
		klines_final: btree.New(degree),
		mutex:        sync.Mutex{},
		degree:       degree,
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
