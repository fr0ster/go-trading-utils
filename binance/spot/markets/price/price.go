package price

import (
	"context"

	"github.com/adshao/go-binance/v2"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	PriceChangeStat binance.PriceChangeStats
	SymbolTicker    binance.SymbolTicker
)

// Less implements btree.Item.
func (p *PriceChangeStat) Less(than btree.Item) bool {
	return p.Symbol < than.(*PriceChangeStat).Symbol
}

// Equal implements btree.Item.
func (p *PriceChangeStat) Equal(than btree.Item) bool {
	return p.Symbol == than.(*PriceChangeStat).Symbol
}

// Equal implements btree.Item.
func (p *SymbolTicker) Less(than btree.Item) bool {
	return p.Symbol < than.(*SymbolTicker).Symbol
}

// Equal implements btree.Item.
func (p *SymbolTicker) Equal(than btree.Item) bool {
	return p.Symbol == than.(*SymbolTicker).Symbol
}

func Init24h(prc *price_types.PriceChangeStats, client *binance.Client, symbols ...string) error {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	pcss, _ :=
		client.NewListPriceChangeStatsService().Symbols(symbols).Do(context.Background())
	for _, pcs := range pcss {
		price, err := Binance2PriceChangeStats(pcs)
		if err != nil {
			return err
		}
		prc.Set(price)
	}
	return nil
}

func Init(prc *price_types.PriceChangeStats, client *binance.Client, symbols ...string) error {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	pcss, _ :=
		client.NewListSymbolTickerService().WindowSize("1m").Symbols(symbols).Do(context.Background())
	for _, pcs := range pcss {
		price, err := Binance2SymbolTicker(pcs)
		if err != nil {
			return err
		}
		prc.Set(price)
	}
	return nil
}

func Binance2PriceChangeStats(binancePriceChangeStats interface{}) (*PriceChangeStat, error) {
	var val PriceChangeStat
	err := copier.Copy(&val, binancePriceChangeStats)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func Binance2SymbolTicker(binanceSymbolTicker interface{}) (*SymbolTicker, error) {
	var val SymbolTicker
	err := copier.Copy(&val, binanceSymbolTicker)
	if err != nil {
		return nil, err
	}
	return &val, nil
}
