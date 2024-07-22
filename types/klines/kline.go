package kline

import (
	"sync"
	"time"

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
		symbolname       string
		interval         KlineStreamInterval
		klines_final     *btree.BTree
		last_kline       *Kline
		mutex            sync.Mutex
		degree           int
		timeOut          time.Duration
		stop             chan struct{}
		resetEvent       chan error
		startKlineStream func() (chan struct{}, chan struct{}, error)
		init             func() error
	}
)

// Kline - тип для зберігання свічок
func (i *Kline) Less(than btree.Item) bool {
	return i.OpenTime < than.(*Kline).OpenTime || i.CloseTime < than.(*Kline).CloseTime
}

func (i *Kline) Equal(than btree.Item) bool {
	return i.OpenTime == than.(*Kline).OpenTime && i.CloseTime == than.(*Kline).CloseTime
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

// GetSymbolname implements kline_interface.Klines.
func (d *Klines) GetSymbolname() string {
	return d.symbolname
}

// SetItem implements kline_interface.Klines.
func (d *Klines) SetKline(value *Kline) {
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

// GetInterval implements kline_interface.Klines.
func (d *Klines) GetInterval() KlineStreamInterval {
	return d.interval
}

// SetInterval implements kline_interface.Klines.
func (d *Klines) SetInterval(interval KlineStreamInterval) {
	d.interval = interval
}

// Kline - B-дерево для зберігання стакана заявок
func New(
	degree int,
	interval KlineStreamInterval,
	symbolname string,
	startKlineStream func(*Klines) func() (chan struct{}, chan struct{}, error),
	initCreator func(*Klines) func() error) *Klines {
	this := &Klines{
		symbolname:   symbolname,
		interval:     interval,
		klines_final: btree.New(degree),
		mutex:        sync.Mutex{},
		degree:       degree,
		timeOut:      0,
	}
	if startKlineStream != nil {
		this.startKlineStream = startKlineStream(this)
	}
	if initCreator != nil {
		this.init = initCreator(this)
		this.init()
	}
	return this
}

func Binance2kline(binanceKline interface{}) (*Kline, error) {
	var val Kline
	err := copier.Copy(&val, binanceKline)
	if err != nil {
		return nil, err
	}

	return &val, nil
}
