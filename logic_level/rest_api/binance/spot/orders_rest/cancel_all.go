package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	common "github.com/fr0ster/turbo-restler/rest_api"

	"github.com/bitly/go-simplejson"
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
	CancelOpenOrdersResponse struct {
		Symbol                  string `json:"symbol"`
		OrigClientOrderID       string `json:"origClientOrderId"`
		OrderID                 int64  `json:"orderId"`
		OrderListID             int64  `json:"orderListId"`
		ClientOrderID           string `json:"clientOrderId"`
		TransactTime            int64  `json:"transactTime"`
		Price                   string `json:"price"`
		OrigQty                 string `json:"origQty"`
		ExecutedQty             string `json:"executedQty"`
		CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
		Status                  string `json:"status"`
		TimeInForce             string `json:"timeInForce"`
		Type                    string `json:"type"`
		Side                    string `json:"side"`
		SelfTradePreventionMode string `json:"selfTradePreventionMode"`
	}
)

// Функція для відміни ордера
func (o *Orders) CancelOrders() ([]*CancelOpenOrdersResponse, error) {
	endpoint := common.EndPoint("/api/v3/openOrders")

	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("symbol", o.symbol)

	body, err := o.CallAPI(http.MethodDelete, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var cancelOrderResponse []*CancelOpenOrdersResponse
	err = json.Unmarshal(body, &cancelOrderResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return cancelOrderResponse, nil
}
