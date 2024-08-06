package spot_web_api

import (
	"encoding/json"
	"fmt"
	"time"

	common "github.com/fr0ster/go-trading-utils/low_level/common"
	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
)

// Структура для параметрів запиту
type (
	OrderParams struct {
		ApiKey           string `json:"apiKey"`
		NewOrderRespType string `json:"newOrderRespType"`
		Price            string `json:"price"`
		Quantity         string `json:"quantity"`
		RecvWindow       int    `json:"recvWindow"`
		Side             string `json:"side"`
		Symbol           string `json:"symbol"`
		TimeInForce      string `json:"timeInForce"`
		Timestamp        int64  `json:"timestamp"`
		Type             string `json:"type"`
		Signature        string `json:"signature"`
	}

	OrderRequest struct {
		ID     string      `json:"id"`
		Method string      `json:"method"`
		Params OrderParams `json:"params"`
	}
	OrderResponse struct {
		ClientOrderId string `json:"clientOrderId"`
		OrderId       int    `json:"orderId"`
		OrderListId   int    `json:"orderListId"`
		Symbol        string `json:"symbol"`
		TransactTime  int64  `json:"transactTime"`
	}
)

// Функція для розміщення ордера через WebSocket
func (wa *WebApi) PlaceOrder(side, orderType, timeInForce, price, quantity string) (response *OrderResponse, limits []web_api.RateLimit, err error) {
	method := "order.place"
	// Створення параметрів запиту
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	// Перетворення структури в строку

	params := OrderParams{
		Symbol:           wa.symbol,
		NewOrderRespType: "ACK",
		Side:             side,
		Type:             orderType,
		TimeInForce:      timeInForce,
		Price:            price,
		RecvWindow:       5000,
		Quantity:         quantity,
		ApiKey:           wa.apiKey,
		Timestamp:        timestamp,
	}
	message, err := common.StructToQueryString(params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	params.Signature = wa.sign.CreateSignature(message)

	msg, limits, err := web_api.CallWebAPI(wa.waHost, wa.waPath, method, params)
	if err != nil {
		return
	}

	err = json.Unmarshal(msg, &response)
	if err != nil {
		return
	}

	return
}
