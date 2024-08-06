package spot_web_api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"

	"github.com/google/uuid"
)

// Структура для параметрів запиту
//
//	type OrderParams struct {
//		ApiKey           string `json:"apiKey"`
//		NewOrderRespType string `json:"newOrderRespType"`
//		Price            string `json:"price"`
//		Quantity         string `json:"quantity"`
//		RecvWindow       int    `json:"recvWindow"`
//		Side             string `json:"side"`
//		Symbol           string `json:"symbol"`
//		TimeInForce      string `json:"timeInForce"`
//		Timestamp        int64  `json:"timestamp"`
//		Type             string `json:"type"`
//	}
type OrderParams struct {
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

type OrderRequest struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params OrderParams `json:"params"`
}

// Функція для створення підпису
func createSignature(secret, message string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// Функція для розміщення ордера через WebSocket
func (wa *WebApi) PlaceOrder(side, orderType, timeInForce, price, quantity string) (response []byte, err error) {
	// Створення параметрів запиту
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	message :=
		"apiKey=" + wa.apiKey +
			"&newOrderRespType=ACK&price=" + price +
			"&quantity=" + quantity +
			"&recvWindow=5000" +
			"&side=" + side +
			"&symbol=" + wa.symbol +
			"&timeInForce=" + timeInForce +
			"&timestamp=" +
			fmt.Sprintf("%d", timestamp) +
			"&type=" + orderType
	signature := createSignature(wa.apiSecret, message)

	params := OrderParams{
		Symbol:      wa.symbol,
		Side:        side,
		Type:        orderType,
		TimeInForce: timeInForce,
		Price:       price,
		Quantity:    quantity,
		ApiKey:      wa.apiKey,
		Timestamp:   timestamp,
		Signature:   signature,
	}

	request := OrderRequest{
		ID:     uuid.New().String(),
		Method: "order.place",
		Params: params,
	}

	// Серіалізація запиту в JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("error marshaling request: %v", err)
		return
	}

	return web_api.CallWebAPI(wa.waHost, wa.waPath, requestBody)
}
