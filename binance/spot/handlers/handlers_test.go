package handlers_test

import (
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"

	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_kline "github.com/fr0ster/go-trading-utils/binance/spot/markets/kline"

	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

const (
	LastUpdateID = int64(2369068)
)

func TestChangingOfOrdersHandler(t *testing.T) {
	even := &binance.WsUserDataEvent{
		Event: binance.UserDataEventTypeExecutionReport,
		OrderUpdate: binance.WsOrderUpdate{
			Status: string(binance.OrderStatusTypeFilled),
		},
	}
	inChannel := make(chan *binance.WsUserDataEvent, 1)
	outChannel :=
		spot_handlers.GetChangingOfOrdersGuard(
			inChannel,
			binance.OrderStatusTypeFilled,
			binance.OrderStatusTypePartiallyFilled)
	inChannel <- even
	res := false
	for {
		select {
		case <-outChannel:
			res = true
		case <-time.After(1000 * time.Millisecond):
			res = false
		}
		if !res {
			t.Fatal("Error sending order event to channel")
		} else {
			break
		}
	}
}

func TestBookTickersUpdateHandler(t *testing.T) {
	even := &binance.WsBookTickerEvent{
		UpdateID:     1,
		Symbol:       "BTCUSDT",
		BestBidPrice: "10000.0",
		BestBidQty:   "210.0",
		BestAskPrice: "11000.0",
		BestAskQty:   "320.0",
	}
	inChannel := make(chan *binance.WsBookTickerEvent, 1)
	bookTicker := bookticker_types.New(3)
	bookTicker.Set(&bookticker_types.BookTicker{
		UpdateID:    0,
		Symbol:      "BTCUSDT",
		BidPrice:    0.0,
		BidQuantity: 0.0,
		AskPrice:    0.0,
		AskQuantity: 0.0,
	})
	outChannel := spot_handlers.GetBookTickersUpdateGuard(bookTicker, inChannel)
	inChannel <- even
	res := false
	for {
		select {
		case <-outChannel:
			res = true
		case <-time.After(1000 * time.Millisecond):
			res = false
		}
		if !res {
			t.Fatal("Error sending order event to channel")
		} else {
			break
		}
	}
}

func TestKlinesUpdateHandler(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	spot := binance.NewClient(api_key, secret_key)

	even := &binance.WsKlineEvent{
		Event:  "kline",
		Time:   1619260800000,
		Symbol: "BTCUSDT",
		Kline: binance.WsKline{
			StartTime:            1619260800000,
			EndTime:              1619260800000,
			Symbol:               "BTCUSDT",
			Interval:             "1m",
			FirstTradeID:         1,
			LastTradeID:          1,
			Open:                 "10000.0",
			Close:                "11000.0",
			High:                 "12000.0",
			Low:                  "9000.0",
			Volume:               "1000.0",
			TradeNum:             1,
			IsFinal:              true,
			QuoteVolume:          "10000.0",
			ActiveBuyVolume:      "1000.0",
			ActiveBuyQuoteVolume: "10000.0",
		},
	}
	klines := kline_types.New(3, "1m", "BTCUSDT")
	spot_kline.Init(klines, spot)

	inChannel := make(chan *binance.WsKlineEvent, 1)
	outChannel := spot_handlers.GetKlinesUpdateGuard(klines, inChannel, true)
	inChannel <- even
	res := false
	for {
		select {
		case <-outChannel:
			res = true
		case <-time.After(1000 * time.Millisecond):
			res = false
		}
		if !res {
			t.Fatal("Error sending order event to channel")
		} else {
			break
		}
	}
}
