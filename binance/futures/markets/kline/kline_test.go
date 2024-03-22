package kline_test

import (
	"testing"

	kline_interface "github.com/fr0ster/go-trading-utils/interfaces/kline"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

func getTestData() []*kline_types.KlineItem {
	return []*kline_types.KlineItem{
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

func TestKlineInterface(t *testing.T) {
	kline := kline_types.New(2)
	test := func(k kline_interface.Klines) {
	}

	// Create a sample Kline instance
	k := getTestData()
	for _, v := range k {
		kline.Set(v)
	}
	test(kline)
}
