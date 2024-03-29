package price

import (
	"errors"
	"sync"

	// prices_interface "github.com/fr0ster/go-trading-utils/interfaces/prices"
	"github.com/google/btree"
)

type (
	PriceChangeStatsItem struct {
		Symbol             string `json:"symbol"`
		PriceChange        string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		WeightedAvgPrice   string `json:"weightedAvgPrice"`
		PrevClosePrice     string `json:"prevClosePrice"`
		LastPrice          string `json:"lastPrice"`
		LastQty            string `json:"lastQty"`
		BidPrice           string `json:"bidPrice"`
		BidQty             string `json:"bidQty"`
		AskPrice           string `json:"askPrice"`
		AskQty             string `json:"askQty"`
		OpenPrice          string `json:"openPrice"`
		HighPrice          string `json:"highPrice"`
		LowPrice           string `json:"lowPrice"`
		Volume             string `json:"volume"`
		QuoteVolume        string `json:"quoteVolume"`
		OpenTime           int64  `json:"openTime"`
		CloseTime          int64  `json:"closeTime"`
		FirstID            int64  `json:"firstId"`
		LastID             int64  `json:"lastId"`
		Count              int64  `json:"count"`
	}
	PriceChangeStats struct {
		tree   btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Less implements btree.Item.
func (p *PriceChangeStatsItem) Less(than btree.Item) bool {
	return p.OpenTime < than.(*PriceChangeStatsItem).OpenTime
}

func (d *PriceChangeStats) Get(symbol string) btree.Item {
	return d.tree.Get(&PriceChangeStatsItem{Symbol: symbol})
}

func (d *PriceChangeStats) Set(value btree.Item) {
	d.tree.ReplaceOrInsert(value)
}

func (d *PriceChangeStats) Lock() {
	d.mutex.Lock()
}

func (d *PriceChangeStats) Unlock() {
	d.mutex.Unlock()
}

// PriceChangeStats - B-дерево для зберігання Цінових змін
func NewPriceChangeStat(degree int) *PriceChangeStats {
	return &PriceChangeStats{
		tree:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func Binance2PriceChangeStats(binancePriceChangeStats interface{}) (*PriceChangeStatsItem, error) {
	switch binancePriceChangeStats := binancePriceChangeStats.(type) {
	case *PriceChangeStatsItem:
		return binancePriceChangeStats, nil
	}
	return nil, errors.New("it's not a PriceChangeStatsItem")
}
