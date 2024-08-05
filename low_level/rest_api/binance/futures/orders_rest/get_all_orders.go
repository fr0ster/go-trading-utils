package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Функція для отримання масиву всіх спотових ордерів
func (o *Orders) GetAllOrders() ([]QueryOrderResponse, error) {
	endpoint := "/fapi/v1/allOrders"

	// Створення параметрів запиту
	params := url.Values{}
	params.Set("symbol", o.symbol)

	body, err := o.CallAPI(http.MethodGet, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var allOrdersResponse []QueryOrderResponse
	err = json.Unmarshal(body, &allOrdersResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return allOrdersResponse, nil
}
