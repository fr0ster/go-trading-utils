package handlers_test

import (
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/stretchr/testify/assert"

	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

const (
	LastUpdateID = int64(2369068)
)

func TestChangingOfOrdersHandler(t *testing.T) {
	even := &futures.WsUserDataEvent{
		Event: futures.UserDataEventTypeOrderTradeUpdate,
		OrderTradeUpdate: futures.WsOrderTradeUpdate{
			Status: futures.OrderStatusTypeFilled,
		},
	}
	inChannel := make(chan *futures.WsUserDataEvent, 1)
	outChannel :=
		futures_handlers.GetChangingOfOrdersGuard(
			inChannel,
			append([]futures.OrderStatusType{futures.OrderStatusTypeFilled}, futures.OrderStatusTypePartiallyFilled))
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

func TestAccountUpdateHandler(t *testing.T) {
	inChannel := make(chan *futures.WsUserDataEvent, 1)
	api_key := os.Getenv("FUTURE_TEST_BINANCE_API_KEY")
	secret_key := os.Getenv("FUTURE_TEST_BINANCE_SECRET_KEY")
	futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(client, 3, []string{"BTC", "USDT"}, []string{"BTCUSDT"})
	assert.Equal(t, nil, err)

	outChannel := futures_handlers.GetAccountInfoGuard(account, inChannel)
	inChannel <- &futures.WsUserDataEvent{
		Event: futures.UserDataEventTypeAccountUpdate,
		Time:  account.UpdateTime + 100,
		AccountUpdate: futures.WsAccountUpdate{
			Reason: "Deposit",
			Balances: []futures.WsBalance{
				{Asset: "BTC", Balance: "0.0", ChangeBalance: "0.0"},
				{Asset: "USDT", Balance: "0.0", ChangeBalance: "0.0"},
			},
			Positions: []futures.WsPosition{
				{
					Symbol:                    "BTCUSDT",
					Side:                      futures.PositionSideTypeLong,
					Amount:                    "0.0",
					MarginType:                futures.MarginTypeIsolated,
					IsolatedWallet:            "0.0",
					EntryPrice:                "0.0",
					MarkPrice:                 "0.0",
					UnrealizedPnL:             "0.0",
					AccumulatedRealized:       "0.0",
					MaintenanceMarginRequired: "0.0"},
				{
					Symbol:                    "BTCUSDT",
					Side:                      futures.PositionSideTypeShort,
					Amount:                    "0.0",
					MarginType:                futures.MarginTypeIsolated,
					IsolatedWallet:            "0.0",
					EntryPrice:                "0.0",
					MarkPrice:                 "0.0",
					UnrealizedPnL:             "0.0",
					AccumulatedRealized:       "0.0",
					MaintenanceMarginRequired: "0.0",
				},
			},
		},
	}
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
	even := &futures.WsBookTickerEvent{
		Symbol:       "BTCUSDT",
		BestBidPrice: "10000.0",
		BestBidQty:   "210.0",
		BestAskPrice: "11000.0",
		BestAskQty:   "320.0",
	}
	inChannel := make(chan *futures.WsBookTickerEvent, 1)
	bookTicker := bookticker_types.New(3)
	bookTicker.Set(&bookticker_types.BookTickerItem{
		Symbol:      "BTCUSDT",
		BidPrice:    0.0,
		BidQuantity: 0.0,
		AskPrice:    0.0,
		AskQuantity: 0.0,
	})
	outChannel := futures_handlers.GetBookTickersUpdateGuard(bookTicker, inChannel)
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

func TestDepthsUpdaterHandler(t *testing.T) {
	inChannel := make(chan *futures.WsDepthEvent, 1)
	outChannel := futures_handlers.GetDepthsUpdateGuard(getTestDepths(), inChannel)
	go func() {
		for i := 0; i < 10; i++ {
			inChannel <- &futures.WsDepthEvent{
				Event:         "depthUpdate",
				Symbol:        "BTCUSDT",
				FirstUpdateID: LastUpdateID - 1,
				LastUpdateID:  LastUpdateID + 2,
				Bids:          []futures.Bid{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(i), 2)}},
				Asks:          []futures.Ask{{Price: "1.93", Quantity: utils.ConvFloat64ToStr(float64(0), 2)}},
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	res := false
	for {
		select {
		case <-outChannel:
			res = true
		case <-time.After(10000 * time.Millisecond):
			res = false
		}
		if !res {
			t.Fatal("Error sending order event to channel")
		} else {
			break
		}
	}
}
