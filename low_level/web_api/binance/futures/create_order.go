package spot_web_api

import (
	"encoding/json"
	"fmt"
	"time"

	common "github.com/fr0ster/go-trading-utils/low_level/common"
	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
	"github.com/google/uuid"
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
)

// Функція для розміщення ордера через WebSocket
func (wa *WebApi) PlaceOrder(side, orderType, timeInForce, price, quantity string) (response []byte, limits []web_api.RateLimit, err error) {
	// Створення параметрів запиту
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	// message := "symbol=" + wa.symbol + "&side=" + side + "&type=" + orderType + "&timeInForce=" + timeInForce + "&price=" + price + "&quantity=" + quantity + "&timestamp=" + strconv.FormatInt(timestamp, 10)
	// message :=
	// 	"apiKey=" + wa.apiKey +
	// 		"&newOrderRespType=ACK" +
	// 		"&price=" + price +
	// 		"&quantity=" + quantity +
	// 		"&recvWindow=5000" +
	// 		"&side=" + side +
	// 		"&symbol=" + wa.symbol +
	// 		"&timeInForce=" + timeInForce +
	// 		"&timestamp=" +
	// 		fmt.Sprintf("%d", timestamp) +
	// 		"&type=" + orderType
	// Перетворення структури в строку

	params := OrderParams{
		Symbol:      wa.symbol,
		Side:        side,
		Type:        orderType,
		TimeInForce: timeInForce,
		Price:       price,
		Quantity:    quantity,
		ApiKey:      wa.apiKey,
		// Signature:   signature,
		Timestamp: timestamp,
	}
	message, err := common.StructToQueryString(params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	params.Signature = wa.sign.CreateSignature(message)

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
