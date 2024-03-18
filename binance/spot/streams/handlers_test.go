package streams_test

import (
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	"github.com/fr0ster/go-trading-utils/binance/spot/streams"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/utils"
)

func TestGetFilledOrderHandler(t *testing.T) {
	even := &binance.WsUserDataEvent{
		Event: binance.UserDataEventTypeExecutionReport,
		OrderUpdate: binance.WsOrderUpdate{
			Status: string(binance.OrderStatusTypeFilled),
		},
	}
	inChannel := make(chan *binance.WsUserDataEvent, 1)
	outChannel := streams.GetFilledOrdersGuard(inChannel)
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

func TestGetBalanceTreeUpdateHandler(t *testing.T) {
	even := &binance.WsUserDataEvent{
		Event: binance.UserDataEventTypeExecutionReport,
		OrderUpdate: binance.WsOrderUpdate{
			Status: string(binance.OrderStatusTypeFilled),
		},
	}
	inChannel := make(chan *binance.WsUserDataEvent, 1)
	balances := markets.BalanceNew(3, nil)
	balances.SetItem(markets.BalanceItemType{
		Asset:  "BTC",
		Free:   0.0,
		Locked: 0.0,
	})
	outChannel := streams.GetBalancesUpdateGuard(balances, inChannel)
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

func TestGetBookTickersUpdateHandler(t *testing.T) {
	even := &binance.WsBookTickerEvent{
		Symbol:       "BTCUSDT",
		BestBidPrice: "10000.0",
		BestBidQty:   "210.0",
		BestAskPrice: "11000.0",
		BestAskQty:   "320.0",
	}
	inChannel := make(chan *binance.WsBookTickerEvent, 1)
	bookTicker := markets.BookTickerNew(3)
	bookTicker.SetItem(markets.BookTickerItemType{
		Symbol:      "BTCUSDT",
		BidPrice:    0.0,
		BidQuantity: 0.0,
		AskPrice:    0.0,
		AskQuantity: 0.0,
	})
	outChannel := streams.GetBookTickersUpdateGuard(bookTicker, inChannel)
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

func getTestDepths() *depth.DepthBTree {
	testDepthTree := depth.DepthNew(3)
	records := []depth_interface.DepthItemType{
		{Price: 1.92, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 150.2},
		{Price: 1.93, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 155.4}, // local maxima
		{Price: 1.94, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 150.0},
		{Price: 1.941, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 130.4},
		{Price: 1.947, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 172.1},
		{Price: 1.948, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 187.4},
		{Price: 1.949, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 236.1}, // local maxima
		{Price: 1.95, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 189.8},
		{Price: 1.951, AskLastUpdateID: 2369068, AskQuantity: 217.9, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.952, AskLastUpdateID: 2369068, AskQuantity: 179.4, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.953, AskLastUpdateID: 2369068, AskQuantity: 180.9, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.954, AskLastUpdateID: 2369068, AskQuantity: 148.5, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.955, AskLastUpdateID: 2369068, AskQuantity: 120.0, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.956, AskLastUpdateID: 2369068, AskQuantity: 110.0, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.957, AskLastUpdateID: 2369068, AskQuantity: 140.0, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.958, AskLastUpdateID: 2369068, AskQuantity: 90.0, BidLastUpdateID: 0, BidQuantity: 0},
	}
	for _, record := range records {
		testDepthTree.ReplaceOrInsert(&record)
	}

	return testDepthTree
}

func TestGetDepthsUpdaterHandler(t *testing.T) {
	inChannel := make(chan *binance.WsDepthEvent, 1)
	outChannel := streams.GetDepthsUpdateGuard(getTestDepths(), inChannel)
	go func() {
		for i := 0; i < 10; i++ {
			inChannel <- &binance.WsDepthEvent{
				Event:         "depthUpdate",
				Symbol:        "BTCUSDT",
				FirstUpdateID: 2369068,
				LastUpdateID:  2369068,
				Bids:          []binance.Bid{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(i), 2)}},
				Asks:          []binance.Ask{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(0), 2)}},
			}
		}
	}()
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
