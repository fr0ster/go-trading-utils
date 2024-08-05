package spot_web_api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

	// Підключення до WebSocket
	u := url.URL{Scheme: "wss", Host: "ws-api.binance.com", Path: "/ws-api/v3"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		err = fmt.Errorf("error connecting to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Відправка запиту на розміщення ордера
	err = conn.WriteMessage(websocket.TextMessage, requestBody)
	if err != nil {
		err = fmt.Errorf("error sending message: %v", err)
		return
	}

	// Читання відповіді
	_, response, err = conn.ReadMessage()
	return
}
