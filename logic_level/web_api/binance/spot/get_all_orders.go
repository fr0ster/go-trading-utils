package spot_web_api

import (
	"fmt"

	common "github.com/fr0ster/turbo-restler/utils/json"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	QueryAllOrdersParams struct {
		ApiKey     string `json:"apiKey"`     // YES
		RecvWindow int    `json:"recvWindow"` // NO: The value cannot be greater than 60000
		Signature  string `json:"signature"`  // YES
		Symbol     string `json:"symbol"`     // NO: If omitted, open orders for all symbols are returned
		Timestamp  int64  `json:"timestamp"`  // YES
	}
	QueryAllOrders struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *QueryOpenOrdersParams
	}

	QueryALlOrdersResults []QueryOpenOrdersResult

	QueryAllOrdersResult struct {
		ClientOrderId           string `json:"clientOrderId"`
		CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
		ExecutedQty             string `json:"executedQty"`
		IcebergQty              string `json:"icebergQty"`
		IsWorking               bool   `json:"isWorking"`
		OrderId                 int64  `json:"orderId"`
		OrderListId             int    `json:"orderListId"`
		OrigQty                 string `json:"origQty"`
		OrigQuoteOrderQty       string `json:"origQuoteOrderQty"`
		Price                   string `json:"price"`
		SelfTradePreventionMode string `json:"selfTradePreventionMode"`
		Side                    string `json:"side"`
		Status                  string `json:"status"`
		StopPrice               string `json:"stopPrice"`
		Symbol                  string `json:"symbol"`
		Time                    int64  `json:"time"`
		TimeInForce             string `json:"timeInForce"`
		Type                    string `json:"type"`
		UpdateTime              int64  `json:"updateTime"`
		WorkingTime             int64  `json:"workingTime"`
	}
)

// Функція для встановлення RecvWindow
func (qoo *QueryAllOrders) SetRecvWindow(recvWindow int) *QueryAllOrders {
	qoo.params.RecvWindow = recvWindow
	return qoo
}

// Функція для розміщення ордера через WebSocket
func (qoo *QueryAllOrders) Do(side, orderType, timeInForce, price, quantity string) (result *QueryOpenOrdersResults, err error) {
	// Перетворення структури в строку
	params, err := common.StructToUrlValues(qoo.params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response, err := web_api.CallWebAPI(qoo.waHost, qoo.waPath, qoo.method, params, qoo.sign)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*QueryOpenOrdersResults)
	return
}

func newQueryAllOrders(apiKey, symbol, waHost, waPath string, sign signature.Sign) *QueryAllOrders {
	return &QueryAllOrders{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "openOrders.status",
		params: &QueryOpenOrdersParams{
			ApiKey: apiKey,
			Symbol: symbol,
		},
	}
}
