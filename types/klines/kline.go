package kline

import (
	"sync"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
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
		startKlineStream types.StreamFunction
		init             types.InitFunction
	}
)

// Kline - тип для зберігання свічок
func (kl *Kline) Less(than btree.Item) bool {
	return kl.OpenTime < than.(*Kline).OpenTime || kl.CloseTime < than.(*Kline).CloseTime
}

func (kl *Kline) Equal(than btree.Item) bool {
	return kl.OpenTime == than.(*Kline).OpenTime && kl.CloseTime == than.(*Kline).CloseTime
}

func (kl *Klines) Ascend(f func(btree.Item) bool) {
	kl.klines_final.Ascend(f)
}

func (kl *Klines) Descend(f func(btree.Item) bool) {
	kl.klines_final.Descend(f)
}

// Lock implements kline_interface.Klines.
func (kl *Klines) Lock() {
	kl.mutex.Lock()
}

// Unlock implements kline_interface.Klines.
func (kl *Klines) Unlock() {
	kl.mutex.Unlock()
}

// GetSymbolname implements kline_interface.Klines.
func (kl *Klines) GetSymbolname() string {
	return kl.symbolname
}

// SetItem implements kline_interface.Klines.
func (kl *Klines) SetKline(value *Kline) {
	kl.klines_final.ReplaceOrInsert(value)
}

// GetLastKline implements kline_interface.Klines.
func (kl *Klines) GetLastKline() *Kline {
	return kl.last_kline
}

// SetLastKline implements kline_interface.Klines.
func (kl *Klines) SetLastKline(value *Kline) {
	kl.last_kline = value
}

// GetKlines implements kline_interface.Klines.
func (kl *Klines) GetKlines() *btree.BTree {
	return kl.klines_final
}

// GetInterval implements kline_interface.Klines.
func (kl *Klines) GetInterval() KlineStreamInterval {
	return kl.interval
}

// SetInterval implements kline_interface.Klines.
func (kl *Klines) SetInterval(interval KlineStreamInterval) {
	kl.interval = interval
}

func (kl *Klines) ResetEvent(err error) {
	kl.resetEvent <- err
}

// Kline - B-дерево для зберігання стакана заявок
func New(
	stop chan struct{},
	degree int,
	interval KlineStreamInterval,
	symbolname string,
	startKlineStream func(*Klines) types.StreamFunction,
	initCreator func(*Klines) types.InitFunction) *Klines {
	this := &Klines{
		symbolname:   symbolname,
		interval:     interval,
		klines_final: btree.New(degree),
		mutex:        sync.Mutex{},
		degree:       degree,
		timeOut:      0,
		stop:         stop,
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
