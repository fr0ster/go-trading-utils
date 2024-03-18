package streams_test

import (
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	"github.com/fr0ster/go-trading-utils/binance/futures/streams"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/utils"
)

func TestGetFilledOrderHandler(t *testing.T) {
	even := &futures.WsUserDataEvent{
		Event: futures.UserDataEventTypeOrderTradeUpdate,
		OrderTradeUpdate: futures.WsOrderTradeUpdate{
			Status: futures.OrderStatusTypeFilled,
		},
	}
	inChannel := make(chan *futures.WsUserDataEvent, 1)
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
	even := &futures.WsUserDataEvent{
		Event: futures.UserDataEventTypeAccountUpdate,
		OrderTradeUpdate: futures.WsOrderTradeUpdate{
			Status: futures.OrderStatusTypeFilled,
		},
	}
	inChannel := make(chan *futures.WsUserDataEvent, 1)
	balances := markets.BalanceNew(3, nil)
	balances.SetItem(markets.BalanceItemType{
		Asset:              "BTC",
		Balance:            0.0,
		ChangeBalance:      0.0,
		CrossWalletBalance: 0.0,
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
	even := &futures.WsBookTickerEvent{
		Symbol:       "BTCUSDT",
		BestBidPrice: "10000.0",
		BestBidQty:   "210.0",
		BestAskPrice: "11000.0",
		BestAskQty:   "320.0",
	}
	inChannel := make(chan *futures.WsBookTickerEvent, 1)
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
	testDepthTree := depth.New(3, 3)
	records := []depth_interface.DepthItemType{
		{Price: 1.92, AskQuantity: 0, BidQuantity: 150.2},
		{Price: 1.93, AskQuantity: 0, BidQuantity: 155.4}, // local maxima
		{Price: 1.94, AskQuantity: 0, BidQuantity: 150.0},
		{Price: 1.941, AskQuantity: 0, BidQuantity: 130.4},
		{Price: 1.947, AskQuantity: 0, BidQuantity: 172.1},
		{Price: 1.948, AskQuantity: 0, BidQuantity: 187.4},
		{Price: 1.949, AskQuantity: 0, BidQuantity: 236.1}, // local maxima
		{Price: 1.95, AskQuantity: 0, BidQuantity: 189.8},
		{Price: 1.951, AskQuantity: 217.9, BidQuantity: 0}, // local maxima
		{Price: 1.952, AskQuantity: 179.4, BidQuantity: 0},
		{Price: 1.953, AskQuantity: 180.9, BidQuantity: 0}, // local maxima
		{Price: 1.954, AskQuantity: 148.5, BidQuantity: 0},
		{Price: 1.955, AskQuantity: 120.0, BidQuantity: 0},
		{Price: 1.956, AskQuantity: 110.0, BidQuantity: 0},
		{Price: 1.957, AskQuantity: 140.0, BidQuantity: 0}, // local maxima
		{Price: 1.958, AskQuantity: 90.0, BidQuantity: 0},
	}
	for _, record := range records {
		testDepthTree.ReplaceOrInsert(&record)
	}

	return testDepthTree
}

func TestGetDepthsUpdaterHandler(t *testing.T) {
	inChannel := make(chan *futures.WsDepthEvent, 1)
	outChannel := streams.GetDepthsUpdateGuard(getTestDepths(), inChannel)
	go func() {
		for i := 0; i < 10; i++ {
			inChannel <- &futures.WsDepthEvent{
				Event:         "depthUpdate",
				Symbol:        "BTCUSDT",
				FirstUpdateID: 2369068,
				LastUpdateID:  2369068,
				Bids:          []futures.Bid{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(i), 2)}},
				Asks:          []futures.Ask{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(0), 2)}},
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
