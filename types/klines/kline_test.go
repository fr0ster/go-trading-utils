package kline_test

import (
	"testing"

	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
)

var (
	quit = make(chan struct{})
)

func getTestData() []*kline_types.Kline {
	return []*kline_types.Kline{
		{
			OpenTime:                 1625097599,
			Open:                     "100",
			High:                     "150",
			Low:                      "80",
			Close:                    "120",
			Volume:                   "1000",
			CloseTime:                1625183999,
			QuoteAssetVolume:         "120000",
			TradeNum:                 100,
			TakerBuyBaseAssetVolume:  "500",
			TakerBuyQuoteAssetVolume: "60000",
		},
		{
			OpenTime:                 1625097600,
			Open:                     "100",
			High:                     "150",
			Low:                      "80",
			Close:                    "120",
			Volume:                   "1000",
			CloseTime:                1625183999,
			QuoteAssetVolume:         "120000",
			TradeNum:                 100,
			TakerBuyBaseAssetVolume:  "500",
			TakerBuyQuoteAssetVolume: "60000",
		},
		{
			OpenTime:                 1625097601,
			Open:                     "100",
			High:                     "150",
			Low:                      "80",
			Close:                    "120",
			Volume:                   "1000",
			CloseTime:                1625183999,
			QuoteAssetVolume:         "120000",
			TradeNum:                 100,
			TakerBuyBaseAssetVolume:  "500",
			TakerBuyQuoteAssetVolume: "60000",
		},
	}
}

func TestKline(t *testing.T) {
	kline := kline_types.New(quit, 2, kline_types.KlineStreamInterval1m, "BTCUSDT", nil, nil)

	test := func(k *kline_types.Klines) {
	}

	// Create a sample Kline instance
	k := getTestData()
	for _, v := range k {
		kline.SetKline(v)
	}
	test(kline)
}
