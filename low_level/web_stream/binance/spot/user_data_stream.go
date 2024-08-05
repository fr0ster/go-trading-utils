package spot_api

import (
	"encoding/json"
	"fmt"

	spot_api "github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/spot"
	common "github.com/fr0ster/go-trading-utils/low_level/web_stream/common"

	"github.com/sirupsen/logrus"
)

// Визначення структури для JSON OrderTradeUpdate
type OrderTradeUpdate struct {
	EventType              string  `json:"e"`
	EventTime              int64   `json:"E"`
	Symbol                 string  `json:"s"`
	ClientOrderID          string  `json:"c"`
	Side                   string  `json:"S"`
	OrderType              string  `json:"o"`
	TimeInForce            string  `json:"f"`
	Quantity               string  `json:"q"`
	Price                  string  `json:"p"`
	StopPrice              string  `json:"P"`
	IcebergQuantity        string  `json:"F"`
	OrderListID            int     `json:"g"`
	OriginalClientOrderID  string  `json:"C"`
	CurrentExecutionType   string  `json:"x"`
	CurrentOrderStatus     string  `json:"X"`
	OrderRejectReason      string  `json:"r"`
	OrderID                int     `json:"i"`
	LastExecutedQuantity   string  `json:"l"`
	CumulativeFilledQty    string  `json:"z"`
	LastExecutedPrice      string  `json:"L"`
	CommissionAmount       string  `json:"n"`
	CommissionAsset        *string `json:"N"`
	TransactionTime        int64   `json:"T"`
	TradeID                int     `json:"t"`
	Ignore                 int     `json:"I"`
	IsOrderWorking         bool    `json:"w"`
	IsBuyerMaker           bool    `json:"m"`
	IsReduceOnly           bool    `json:"M"`
	OrderCreationTime      int64   `json:"O"`
	CumulativeQuoteQty     string  `json:"Z"`
	LastQuoteAssetTransact string  `json:"Y"`
	QuoteOrderQty          string  `json:"Q"`
	WorkingTime            int64   `json:"W"`
	SelfTradePrevention    string  `json:"V"`
}

// Функція для парсингу JSON
func parseJSON(data []byte) (*OrderTradeUpdate, error) {
	var orderTradeUpdate OrderTradeUpdate
	err := json.Unmarshal(data, &orderTradeUpdate)
	if err != nil {
		return nil, err
	}
	return &orderTradeUpdate, nil
}

func UserDataStream(apiKey string, symbol string, callBack func(*OrderTradeUpdate), quit chan struct{}, useTestNet ...bool) {
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
