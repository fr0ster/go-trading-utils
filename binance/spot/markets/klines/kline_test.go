package kline_test

import (
	"testing"

	kline_interface "github.com/fr0ster/go-trading-utils/interfaces/kline"
	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
	"github.com/stretchr/testify/assert"
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

func TestKlineInterface(t *testing.T) {
	klines := kline_types.New(quit, 2, "1m", "BTCUSDT", nil, nil)

	test := func(k kline_interface.Klines) {
		val1 := k.GetKlines().Max().(*kline_types.Kline)
		assert.Equal(t, int64(1625097601), val1.OpenTime)
		val2 := k.GetKlines().Min().(*kline_types.Kline)
		assert.Equal(t, int64(1625097599), val2.OpenTime)
	}

	// Create a sample Kline instance
	k := getTestData()
	for _, v := range k {
		klines.SetKline(v)
	}
	test(klines)
}
