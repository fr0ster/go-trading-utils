package price

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	PriceChangeStat futures.PriceChangeStats
	SymbolPrice     futures.SymbolPrice
)

// Less is a method for implementing the btree.Item interface.
func (pcs *PriceChangeStat) Less(item btree.Item) bool {
	return pcs.Symbol < item.(*PriceChangeStat).Symbol
}

// Equal is a method for implementing the btree.Item interface.
func (pcs *PriceChangeStat) Equal(item btree.Item) bool {
	return pcs.Symbol == item.(*PriceChangeStat).Symbol
}

// Less is a method for implementing the btree.Item interface.
func (pcs *SymbolPrice) Less(item btree.Item) bool {
	return pcs.Symbol < item.(*SymbolPrice).Symbol
}

// Equal is a method for implementing the btree.Item interface.
func (pcs *SymbolPrice) Equal(item btree.Item) bool {
	return pcs.Symbol == item.(*SymbolPrice).Symbol
}

func Init24h(prc *price_types.PriceChangeStats, client futures.Client, symbols ...string) (err error) {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	var pcss []*futures.PriceChangeStats
	if len(symbols) > 0 {
		for _, symbol := range symbols {
			res, _ :=
				client.NewListPriceChangeStatsService().Symbol(symbol).Do(context.Background())
			pcss = append(pcss, res...)
		}
	} else {
		pcss, err =
			client.NewListPriceChangeStatsService().Do(context.Background())
		if err != nil {
			return err
		}
	}
	for _, pcs := range pcss {
		prc.Set(&PriceChangeStat{
			Symbol:             pcs.Symbol,
			PriceChange:        pcs.PriceChange,
			PriceChangePercent: pcs.PriceChangePercent,
			WeightedAvgPrice:   pcs.WeightedAvgPrice,
			PrevClosePrice:     pcs.PrevClosePrice,
			LastPrice:          pcs.LastPrice,
			OpenPrice:          pcs.OpenPrice,
			HighPrice:          pcs.HighPrice,
			LowPrice:           pcs.LowPrice,
			Volume:             pcs.Volume,
			QuoteVolume:        pcs.QuoteVolume,
			OpenTime:           pcs.OpenTime,
			CloseTime:          pcs.CloseTime,
			FirstID:            pcs.FirstID,
			LastID:             pcs.LastID,
			Count:              pcs.Count,
		})
	}
	return nil
}

func Init(prc *price_types.PriceChangeStats, client futures.Client, symbols ...string) (err error) {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	var pcss []*futures.SymbolPrice
	if len(symbols) > 0 {
		for _, symbol := range symbols {
			res, _ :=
				client.NewListPricesService().Symbol(symbol).Do(context.Background())
			pcss = append(pcss, res...)
		}
	} else {
		pcss, err =
			client.NewListPricesService().Do(context.Background())
		if err != nil {
			return err
		}
	}
	for _, pcs := range pcss {
		prc.Set(&SymbolPrice{
			Symbol: pcs.Symbol,
			Price:  pcs.Price,
		})
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
