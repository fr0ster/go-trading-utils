package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Структура для параметрів запиту
type SpotOrderUpdateRequest struct {
	Symbol    string  `json:"symbol"`
	OrderID   int64   `json:"orderId"`
	Quantity  float64 `json:"quantity,omitempty"`
	Price     float64 `json:"price,omitempty"`
	Timestamp int64   `json:"timestamp"`
	Signature string  `json:"signature"`
}

// Структура для відповіді API
type SpotOrderUpdateResponse struct {
	OrderID int64  `json:"orderId"`
	Symbol  string `json:"symbol"`
	Status  string `json:"status"`
}

// Функція для зміни ордера
func (o *Orders) UpdateSpotOrder(orderID int64, quantity, price float64) (*SpotOrderUpdateResponse, error) {
	endpoint := "/api/v3/order"

	// Створення параметрів запиту
	params := url.Values{}
	params.Set("symbol", o.symbol)
	params.Set("orderId", strconv.FormatInt(orderID, 10))
	if quantity > 0 {
		params.Set("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
	}
	if price > 0 {
		params.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	}

	body, err := o.CallAPI(http.MethodPut, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var spotOrderUpdateResponse SpotOrderUpdateResponse
	err = json.Unmarshal(body, &spotOrderUpdateResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &spotOrderUpdateResponse, nil
}
