package kline_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_kline "github.com/fr0ster/go-trading-utils/binance/spot/markets/kline"
	kline_interface "github.com/fr0ster/go-trading-utils/interfaces/kline"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

func getTestData() []*kline_types.Kline {
	return []*kline_types.Kline{
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
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	spot := binance.NewClient(api_key, secret_key)

	kline := kline_types.New(2)
	spot_kline.Init(kline, spot, "BTCUSDT")
	test := func(k kline_interface.Klines) {
	}

	// Create a sample Kline instance
	k := getTestData()
	for _, v := range k {
		kline.Set(v)
	}
	test(kline)
}
