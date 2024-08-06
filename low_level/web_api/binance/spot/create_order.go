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
		Signature:        signature,
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
