package spot_web_api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fr0ster/go-trading-utils/low_level/common"
	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"

	"github.com/google/uuid"
)

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
	// Створення параметрів запиту
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	params := OrderParams{
		ApiKey:           wa.apiKey,
		NewOrderRespType: "ACK",
		Price:            price,
		Quantity:         quantity,
		RecvWindow:       5000,
		Side:             side,
		Symbol:           wa.symbol,
		TimeInForce:      timeInForce,
		Timestamp:        timestamp,
		Type:             orderType,
	}
	message, err := common.StructToQueryString(params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	params.Signature = wa.sign.CreateSignature(message)

	// request := OrderRequest{
	// 	ID:     uuid.New().String(),
	// 	Method: "order.place",
	// 	Params: params,
	// }

	// // Серіалізація запиту в JSON
	// requestBody, err := json.Marshal(request)
	// if err != nil {
	// 	err = fmt.Errorf("error marshaling request: %v", err)
	// 	return
	// }

	// msg, limits, err := web_api.CallWebAPI(wa.waHost, wa.waPath, requestBody)
	// if err != nil {
	// 	return
	// }
	msg, limits, err := wa.callWebApi("order.place", params)
	if err != nil {
		return
	}

	err = json.Unmarshal(msg, &response)
	if err != nil {
		return
	}

	return
}

type Request struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func (wa *WebApi) callWebApi(method string, params interface{}) (response []byte, limits []web_api.RateLimit, err error) {
	request := Request{
		ID:     uuid.New().String(),
		Method: method,
		Params: params,
	}

	// Серіалізація запиту в JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("error marshaling request: %v", err)
		return
	}

	response, limits, err = web_api.CallWebAPI(wa.waHost, wa.waPath, requestBody)

	return
}
