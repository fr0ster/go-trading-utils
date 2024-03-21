package kline_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/markets/kline"
)

func getTestData() []kline.KlineItem {
	return []kline.KlineItem{
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
	// Create a sample Kline instance
	k := &kline.KlineItem{
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
	}

	// Test the Less method
	less := k.Less(&kline.KlineItem{OpenTime: 1625097601})
	if !less {
		t.Errorf("Expected k.Less to be true, got false")
	}

	// Test the Equal method
	equal := k.Equal(&kline.KlineItem{OpenTime: 1625097600})
	if !equal {
		t.Errorf("Expected k.Equal to be true, got false")
	}
}

func TestKlines(t *testing.T) {
	// Create a sample Klines implementation
	kl := kline.New(3)

	// Test the Lock and Unlock methods
	kl.Lock()
	defer kl.Unlock()

	// Test the Init method
	for _, k := range getTestData() {
		kl.Set(&k)
	}

	// Test the GetItem method
	item := kl.Get(1625097600)
	if item == nil {
		t.Errorf("Expected kl.GetItem to return a non-nil value")
	}

	// Test the SetItem method
	kl.Set(&kline.KlineItem{
		OpenTime:                 1625097602,
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
	})
}
