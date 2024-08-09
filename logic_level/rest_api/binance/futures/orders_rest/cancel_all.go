package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitly/go-simplejson"
	common "github.com/fr0ster/turbo-restler/rest_api"
)

type (
	// Структура для параметрів запиту
	AllOrderCancelRequest struct {
		Symbol    string `json:"symbol"`
		OrderID   int64  `json:"orderId,omitempty"`
		Timestamp int64  `json:"timestamp"`
		Signature string `json:"signature"`
	}

	// Структура для відповіді API
	AllOrderCancelResponse struct {
		OrderID int64  `json:"orderId"`
		Symbol  string `json:"symbol"`
		Status  string `json:"status"`
	}
)

// Функція для відміни ордера
func (o *Orders) CancelOrders() (*AllOrderCancelResponse, error) {
	endpoint := common.EndPoint("/fapi/v1/allOpenOrders")

	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("symbol", o.symbol)

	body, err := o.CallAPI(http.MethodDelete, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var cancelOrderResponse AllOrderCancelResponse
	err = json.Unmarshal(body, &cancelOrderResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &cancelOrderResponse, nil
}
