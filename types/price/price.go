package price

import (
	"sync"

	// prices_interface "github.com/fr0ster/go-trading-utils/interfaces/prices"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	PriceChangeStat struct {
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
func (p *PriceChangeStat) Less(than btree.Item) bool {
	return p.OpenTime < than.(*PriceChangeStat).OpenTime
}

func (d *PriceChangeStats) Get(symbol string) btree.Item {
	return d.tree.Get(&PriceChangeStat{Symbol: symbol})
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
func New(degree int) *PriceChangeStats {
	return &PriceChangeStats{
		tree:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func Binance2PriceChangeStats(binancePriceChangeStats interface{}) (*PriceChangeStat, error) {
	var val PriceChangeStat
	err := copier.Copy(&val, binancePriceChangeStats)
	if err != nil {
		return nil, err
	}
	return &val, nil
}
