package handlers_test

import (
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/google/btree"
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
		handlers.GetChangingOfOrdersGuard(
			inChannel,
			binance.UserDataEventTypeExecutionReport,
			append([]binance.OrderStatusType{binance.OrderStatusTypeFilled}, binance.OrderStatusTypeFilled))
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

func TestBalanceTreeUpdateHandler(t *testing.T) {
	even := &binance.WsUserDataEvent{
		Event: binance.UserDataEventTypeExecutionReport,
		OrderUpdate: binance.WsOrderUpdate{
			Status: string(binance.OrderStatusTypeFilled),
		},
	}
	inChannel := make(chan *binance.WsUserDataEvent, 1)
	bt := balances_types.New(3)
	bt.SetItem(&balances_types.BalanceItemType{
		Asset:  "BTC",
		Free:   0.0,
		Locked: 0.0,
	})
	outChannel := handlers.GetBalancesUpdateGuard(bt, inChannel)
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
		Symbol:       "BTCUSDT",
		BestBidPrice: "10000.0",
		BestBidQty:   "210.0",
		BestAskPrice: "11000.0",
		BestAskQty:   "320.0",
	}
	inChannel := make(chan *binance.WsBookTickerEvent, 1)
	bookTicker := bookticker_types.New(3)
	bookTicker.Set(&bookticker_types.BookTickerItem{
		Symbol:      "BTCUSDT",
		BidPrice:    0.0,
		BidQuantity: 0.0,
		AskPrice:    0.0,
		AskQuantity: 0.0,
	})
	outChannel := handlers.GetBookTickersUpdateGuard(bookTicker, inChannel)
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

func getTestDepths() *depth_types.Depth {
	bids := btree.New(3)
	bidList := []depth_types.DepthItemType{
		{Price: 1.92, Quantity: 150.2},
		{Price: 1.93, Quantity: 155.4}, // local maxima
		{Price: 1.94, Quantity: 150.0},
		{Price: 1.941, Quantity: 130.4},
		{Price: 1.947, Quantity: 172.1},
		{Price: 1.948, Quantity: 187.4},
		{Price: 1.949, Quantity: 236.1}, // local maxima
		{Price: 1.95, Quantity: 189.8},
	}
	asks := btree.New(3)
	askList := []depth_types.DepthItemType{
		{Price: 1.951, Quantity: 217.9}, // local maxima
		{Price: 1.952, Quantity: 179.4},
		{Price: 1.953, Quantity: 180.9}, // local maxima
		{Price: 1.954, Quantity: 148.5},
		{Price: 1.955, Quantity: 120.0},
		{Price: 1.956, Quantity: 110.0},
		{Price: 1.957, Quantity: 140.0}, // local maxima
		{Price: 1.958, Quantity: 90.0},
	}
	for _, bid := range bidList {
		bids.ReplaceOrInsert(&bid)
	}
	for _, ask := range askList {
		asks.ReplaceOrInsert(&ask)
	}
	ds := depth_types.NewDepth(3, "SUSHIUSDT")
	ds.LastUpdateID = LastUpdateID
	ds.SetAsks(asks)
	ds.SetBids(bids)

	return ds
}

// func TestDepthsUpdaterHandler(t *testing.T) {
// 	inChannel := make(chan *binance.WsDepthEvent, 1)
// 	outChannel := handlers.GetDepthsUpdateGuard(getTestDepths(), inChannel)
// 	go func() {
// 		for i := 0; i < 10; i++ {
// 			inChannel <- &binance.WsDepthEvent{
// 				Event:         "depthUpdate",
// 				Symbol:        "BTCUSDT",
// 				FirstUpdateID: LastUpdateID - 1,
// 				LastUpdateID:  LastUpdateID + 1,
// 				Bids:          []binance.Bid{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(i), 2)}},
// 				Asks:          []binance.Ask{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(0), 2)}},
// 			}
// 		}
// 	}()
// 	res := false
// 	for {
// 		select {
// 		case <-outChannel:
// 			res = true
// 		case <-time.After(1000 * time.Millisecond):
// 			res = false
// 		}
// 		if !res {
// 			t.Fatal("Error sending order event to channel")
// 		} else {
// 			break
// 		}
// 	}
// }
