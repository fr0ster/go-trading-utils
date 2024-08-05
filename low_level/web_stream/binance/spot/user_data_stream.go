package spot_api

import (
	"encoding/json"
	"fmt"

	spot_api "github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/spot"
	common "github.com/fr0ster/go-trading-utils/low_level/web_stream/common"
	"github.com/fr0ster/go-trading-utils/types"

	"github.com/sirupsen/logrus"
)

// Визначення структури для JSON OrderTradeUpdate
// WsUserDataEvent define user data event
type WsUserDataEvent struct {
	Event         types.UserDataEventType `json:"e"`
	Time          int64                   `json:"E"`
	AccountUpdate WsAccountUpdateList
	BalanceUpdate WsBalanceUpdate
	OrderUpdate   WsOrderUpdate
	OCOUpdate     WsOCOUpdate
}

type WsAccountUpdateList struct {
	AccountUpdateTime int64             `json:"u"`
	WsAccountUpdates  []WsAccountUpdate `json:"B"`
}

// WsAccountUpdate define account update
type WsAccountUpdate struct {
	Asset  string `json:"a"`
	Free   string `json:"f"`
	Locked string `json:"l"`
}

type WsBalanceUpdate struct {
	Asset           string `json:"a"`
	Change          string `json:"d"`
	TransactionTime int64  `json:"T"`
}

type WsOrderUpdate struct {
	Symbol                  string                `json:"s"`
	ClientOrderId           string                `json:"c"`
	Side                    string                `json:"S"`
	Type                    string                `json:"o"`
	TimeInForce             types.TimeInForceType `json:"f"`
	Volume                  string                `json:"q"`
	Price                   string                `json:"p"`
	StopPrice               string                `json:"P"`
	IceBergVolume           string                `json:"F"`
	OrderListId             int64                 `json:"g"` // for OCO
	OrigCustomOrderId       string                `json:"C"` // customized order ID for the original order
	ExecutionType           string                `json:"x"` // execution type for this event NEW/TRADE...
	Status                  string                `json:"X"` // order status
	RejectReason            string                `json:"r"`
	Id                      int64                 `json:"i"` // order id
	LatestVolume            string                `json:"l"` // quantity for the latest trade
	FilledVolume            string                `json:"z"`
	LatestPrice             string                `json:"L"` // price for the latest trade
	FeeAsset                string                `json:"N"`
	FeeCost                 string                `json:"n"`
	TransactionTime         int64                 `json:"T"`
	TradeId                 int64                 `json:"t"`
	IgnoreI                 int64                 `json:"I"` // ignore
	IsInOrderBook           bool                  `json:"w"` // is the order in the order book?
	IsMaker                 bool                  `json:"m"` // is this order maker?
	IgnoreM                 bool                  `json:"M"` // ignore
	CreateTime              int64                 `json:"O"`
	FilledQuoteVolume       string                `json:"Z"` // the quote volume that already filled
	LatestQuoteVolume       string                `json:"Y"` // the quote volume for the latest trade
	QuoteVolume             string                `json:"Q"`
	SelfTradePreventionMode string                `json:"V"`

	//These are fields that appear in the payload only if certain conditions are met.
	TrailingDelta              int64  `json:"d"` // Appears only for trailing stop orders.
	TrailingTime               int64  `json:"D"`
	StrategyId                 int64  `json:"j"` // Appears only if the strategyId parameter was provided upon order placement.
	StrategyType               int64  `json:"J"` // Appears only if the strategyType parameter was provided upon order placement.
	PreventedMatchId           int64  `json:"v"` // Appears only for orders that expired due to STP.
	PreventedQuantity          string `json:"A"`
	LastPreventedQuantity      string `json:"B"`
	TradeGroupId               int64  `json:"u"`
	CounterOrderId             int64  `json:"U"`
	CounterSymbol              string `json:"Cs"`
	PreventedExecutionQuantity string `json:"pl"`
	PreventedExecutionPrice    string `json:"pL"`
	PreventedExecutionQuoteQty string `json:"pY"`
	WorkingTime                int64  `json:"W"` // Appears when the order is working on the book
	MatchType                  string `json:"b"`
	AllocationId               int64  `json:"a"`
	WorkingFloor               string `json:"k"`  // Appears for orders that could potentially have allocations
	UsedSor                    bool   `json:"uS"` // Appears for orders that used SOR
}

type WsOCOUpdate struct {
	Symbol          string `json:"s"`
	OrderListId     int64  `json:"g"`
	ContingencyType string `json:"c"`
	ListStatusType  string `json:"l"`
	ListOrderStatus string `json:"L"`
	RejectReason    string `json:"r"`
	ClientOrderId   string `json:"C"` // List Client Order ID
	TransactionTime int64  `json:"T"`
	Orders          WsOCOOrderList
}

type WsOCOOrderList struct {
	WsOCOOrders []WsOCOOrder `json:"O"`
}

type WsOCOOrder struct {
	Symbol        string `json:"s"`
	OrderId       int64  `json:"i"`
	ClientOrderId string `json:"c"`
}

// Функція для парсингу JSON
func parseJSON(data []byte) (*WsUserDataEvent, error) {
	var orderTradeUpdate WsUserDataEvent
	err := json.Unmarshal(data, &orderTradeUpdate)
	if err != nil {
		return nil, err
	}
	return &orderTradeUpdate, nil
}

func UserDataStream(apiKey string, symbol string, callBack func(*WsUserDataEvent), quit chan struct{}, useTestNet ...bool) {
	wss := GetWsEndpoint(useTestNet...)
	listenKey, err := spot_api.ListenKey(apiKey, useTestNet...)
	if err != nil {
		logrus.Fatalf("Error getting listen key: %v", err)
	}
	wsURL := fmt.Sprintf("%s/%s", wss, listenKey)
	common.StartStreamer(wsURL, func(message []byte) {
		orderTradeUpdate, err := parseJSON(message)
		if err != nil {
			logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
		}
		if callBack != nil {
			callBack(orderTradeUpdate)
		}
	}, quit)
}
