package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
)

type (
	// Структура для параметрів запиту
	OrderCancelRequest struct {
		Symbol    string `json:"symbol"`
		OrderID   int64  `json:"orderId,omitempty"`
		Timestamp int64  `json:"timestamp"`
		Signature string `json:"signature"`
	}

	// Структура для відповіді API
	OrderCancelResponse struct {
		OrderID int64  `json:"orderId"`
		Symbol  string `json:"symbol"`
		Status  string `json:"status"`
	}
)

// Функція для відміни ордера
func (o *Orders) CancelOrder(orderID int64) (*OrderCancelResponse, error) {
	endpoint := "/fapi/v1/order"

	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("symbol", o.symbol)
	params.Set("orderId", strconv.FormatInt(orderID, 10))

	body, err := o.CallAPI(http.MethodDelete, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var cancelOrderResponse OrderCancelResponse
	err = json.Unmarshal(body, &cancelOrderResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &cancelOrderResponse, nil
}
