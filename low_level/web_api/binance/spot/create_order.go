package spot_web_api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"

	"github.com/google/uuid"
)

// Структура для параметрів запиту
type OrderParams struct {
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	Type        string `json:"type"`
	TimeInForce string `json:"timeInForce"`
	Price       string `json:"price"`
	Quantity    string `json:"quantity"`
	APIKey      string `json:"apiKey"`
	Signature   string `json:"signature"`
	Timestamp   int64  `json:"timestamp"`
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
	message := "symbol=" + wa.symbol + "&side=" + side + "&type=" + orderType + "&timeInForce=" + timeInForce + "&price=" + price + "&quantity=" + quantity + "&timestamp=" + strconv.FormatInt(timestamp, 10)
	signature := createSignature(wa.apiSecret, message)

	params := OrderParams{
		Symbol:      wa.symbol,
		Side:        side,
		Type:        orderType,
		TimeInForce: timeInForce,
		Price:       price,
		Quantity:    quantity,
		APIKey:      wa.apiKey,
		Signature:   signature,
		Timestamp:   timestamp,
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
