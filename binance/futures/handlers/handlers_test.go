package handlers_test

import (
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	accounts "github.com/fr0ster/go-trading-utils/binance/futures/markets/account"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/balances"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

func TestGetFilledOrderHandler(t *testing.T) {
	even := &futures.WsUserDataEvent{
		Event: futures.UserDataEventTypeOrderTradeUpdate,
		OrderTradeUpdate: futures.WsOrderTradeUpdate{
			Status: futures.OrderStatusTypeFilled,
		},
	}
	inChannel := make(chan *futures.WsUserDataEvent, 1)
	outChannel := handlers.GetFilledOrdersGuard(inChannel)
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

	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	account, err := accounts.New(futures.NewClient(api_key, secret_key), 3)
	if err != nil || account == nil {
		t.Errorf("Error creating account: %v", err)
	}

	accountAsset := &futures.AccountAsset{
		Asset:                  "BTC",
		InitialMargin:          "0.0",
		MaintMargin:            "0.0",
		MarginBalance:          "0.0",
		MaxWithdrawAmount:      "0.0",
		OpenOrderInitialMargin: "0.0",
		PositionInitialMargin:  "0.0",
		UnrealizedProfit:       "0.0",
		WalletBalance:          "0.0",
		CrossWalletBalance:     "0.0",
		CrossUnPnl:             "0.0",
		AvailableBalance:       "0.0",
		MarginAvailable:        false,
		UpdateTime:             0,
	}
	bt := balances.New(3, append([]*futures.AccountAsset{}, accountAsset))
	bt.SetItem(balances.BalanceItemType{
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

func TestGetBookTickersUpdateHandler(t *testing.T) {
	even := &futures.WsBookTickerEvent{
		Symbol:       "BTCUSDT",
		BestBidPrice: "10000.0",
		BestBidQty:   "210.0",
		BestAskPrice: "11000.0",
		BestAskQty:   "320.0",
	}
	inChannel := make(chan *futures.WsBookTickerEvent, 1)
	bookTicker := bookticker_types.New(3)
	bookTicker.Set(bookticker_types.BookTickerItem{
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

func getTestDepths() *depth.Depth {
	bids := btree.New(3)
	bidList := []depth.DepthItemType{
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
	askList := []depth.DepthItemType{
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
		bids.ReplaceOrInsert(bid)
	}
	for _, ask := range askList {
		asks.ReplaceOrInsert(ask)
	}
	ds := depth.New(3, 2, 5, "SUSHIUSDT")
	ds.SetAsks(asks)
	ds.SetBids(bids)

	return ds
}

func TestGetDepthsUpdaterHandler(t *testing.T) {
	inChannel := make(chan *futures.WsDepthEvent, 1)
	outChannel := handlers.GetDepthsUpdateGuard(getTestDepths(), inChannel)
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
