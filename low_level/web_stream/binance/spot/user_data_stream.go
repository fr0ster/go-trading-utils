package spot_api

import (
	"encoding/json"
	"fmt"

	spot_api "github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/spot"
	types "github.com/fr0ster/go-trading-utils/low_level/web_stream/binance/common"
	common "github.com/fr0ster/go-trading-utils/low_level/web_stream/common"

	"github.com/sirupsen/logrus"
)

// // Визначення структури для JSON
// type OrderTradeUpdate struct {
// 	EventType    string `json:"e"`
// 	EventTime    int64  `json:"E"`
// 	TransactTime int64  `json:"T"`
// 	Order        struct {
// 		Symbol              string `json:"s"`
// 		ClientOrderID       string `json:"c"`
// 		Side                string `json:"S"`
// 		OrderType           string `json:"o"`
// 		TimeInForce         string `json:"f"`
// 		Quantity            string `json:"q"`
// 		Price               string `json:"p"`
// 		AveragePrice        string `json:"ap"`
// 		StopPrice           string `json:"sp"`
// 		ExecutionType       string `json:"x"`
// 		OrderStatus         string `json:"X"`
// 		OrderID             int64  `json:"i"`
// 		LastFilledQuantity  string `json:"l"`
// 		CumulativeQuantity  string `json:"z"`
// 		LastFilledPrice     string `json:"L"`
// 		CommissionAmount    string `json:"n"`
// 		CommissionAsset     string `json:"N"`
// 		TradeTime           int64  `json:"T"`
// 		TradeID             int64  `json:"t"`
// 		BidNotional         string `json:"b"`
// 		AskNotional         string `json:"a"`
// 		IsMaker             bool   `json:"m"`
// 		IsReduceOnly        bool   `json:"R"`
// 		WorkingType         string `json:"wt"`
// 		OriginalOrderType   string `json:"ot"`
// 		PositionSide        string `json:"ps"`
// 		IsClosePosition     bool   `json:"cp"`
// 		RealizedProfit      string `json:"rp"`
// 		IsPriceProtect      bool   `json:"pP"`
// 		StopOrderType       int64  `json:"si"`
// 		StopOrderStatus     int64  `json:"ss"`
// 		ActivationPriceType string `json:"V"`
// 		PriceProtectMode    string `json:"pm"`
// 		GTDTime             int64  `json:"gtd"`
// 	} `json:"o"`
// }

// Функція для парсингу JSON
func parseJSON(data []byte) (*types.OrderTradeUpdate, error) {
	var orderTradeUpdate types.OrderTradeUpdate
	err := json.Unmarshal(data, &orderTradeUpdate)
	if err != nil {
		return nil, err
	}
	return &orderTradeUpdate, nil
}

func UserDataStream(apiKey string, symbol string, callBack func(*types.OrderTradeUpdate), quit chan struct{}, useTestNet ...bool) {
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
