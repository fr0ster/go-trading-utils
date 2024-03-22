package price_test

import (
	"testing"

	prices_interface "github.com/fr0ster/go-trading-utils/interfaces/price"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
)

func getTestData() []*price_types.PriceChangeStatsItem {
	return append([]*price_types.PriceChangeStatsItem{
		{
			Symbol:             "BTCUSDT",
			PriceChange:        "0.00000000",
			PriceChangePercent: "0.000",
			WeightedAvgPrice:   "0.00000000",
			PrevClosePrice:     "0.00000000",
			LastPrice:          "0.00000000",
			LastQty:            "0.00000000",
			BidPrice:           "0.00000000",
			BidQty:             "0.00000000",
			AskPrice:           "0.00000000",
			AskQty:             "0.00000000",
			OpenPrice:          "0.00000000",
			HighPrice:          "0.00000000",
			LowPrice:           "0.00000000",
			Volume:             "0.00000000",
			QuoteVolume:        "0.00000000",
			OpenTime:           0,
			CloseTime:          0,
			FirstID:            0,
			LastID:             0,
			Count:              0,
		},
		{
			Symbol:             "ETHUSDT",
			PriceChange:        "0.00000000",
			PriceChangePercent: "0.000",
			WeightedAvgPrice:   "0.00000000",
			PrevClosePrice:     "0.00000000",
			LastPrice:          "0.00000000",
			LastQty:            "0.00000000",
			BidPrice:           "0.00000000",
			BidQty:             "0.00000000",
			AskPrice:           "0.00000000",
			AskQty:             "0.00000000",
			OpenPrice:          "0.00000000",
			HighPrice:          "0.00000000",
			LowPrice:           "0.00000000",
			Volume:             "0.00000000",
			QuoteVolume:        "0.00000000",
			OpenTime:           0,
			CloseTime:          0,
			FirstID:            0,
			LastID:             0,
			Count:              0,
		},
	}, nil...)
}

func TestPricesInterfaces(t *testing.T) {
	pcs := price_types.NewPriceChangeStat(2)
	test := func(p prices_interface.Prices) {
		p.Lock()
		defer p.Unlock()
		// p.Init("test", "test", "BTCUSDT", true)
		for _, k := range getTestData() {
			p.Set(&price_types.PriceChangeStatsItem{
				Symbol:             k.Symbol,
				PriceChange:        k.PriceChange,
				PriceChangePercent: k.PriceChangePercent,
				WeightedAvgPrice:   k.WeightedAvgPrice,
				PrevClosePrice:     k.PrevClosePrice,
				LastPrice:          k.LastPrice,
				LastQty:            k.LastQty,
				BidPrice:           k.BidPrice,
				BidQty:             k.BidQty,
				AskPrice:           k.AskPrice,
				AskQty:             k.AskQty,
				OpenPrice:          k.OpenPrice,
				HighPrice:          k.HighPrice,
				LowPrice:           k.LowPrice,
				Volume:             k.Volume,
				QuoteVolume:        k.QuoteVolume,
				OpenTime:           k.OpenTime,
				CloseTime:          k.CloseTime,
				FirstID:            k.FirstID,
				LastID:             k.LastID,
				Count:              k.Count,
			})
		}
		for _, k := range getTestData() {
			if p.Get(k.Symbol) == nil {
				t.Error("Expected to find PriceChangeStatsItem")
			}
		}
	}
	test(pcs)
}
