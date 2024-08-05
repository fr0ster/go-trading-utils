package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Функція для створення ордера
func (o *Orders) CreateOrder(side, orderType string, quantity, price float64, timeInForce string) (*CreateOrderResponse, error) {
	endpoint := "/fapi/v1/order"

	// Створення параметрів запиту
	params := url.Values{}
	params.Set("symbol", o.symbol)
	params.Set("side", side)
	params.Set("type", orderType)
	params.Set("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
	if price > 0 {
		params.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	}
	if timeInForce != "" {
		params.Set("timeInForce", timeInForce)
	}

	body, err := o.CallAPI(http.MethodPost, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var spotOrderResponse CreateOrderResponse
	err = json.Unmarshal(body, &spotOrderResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &spotOrderResponse, nil
}
