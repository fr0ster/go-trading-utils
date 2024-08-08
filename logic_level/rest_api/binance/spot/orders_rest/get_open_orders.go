package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitly/go-simplejson"
)

// Функція для отримання відкритих спотових ордерів
func (o *Orders) GetOpenOrders() ([]QueryOrderResponse, error) {
	endpoint := "/api/v3/openOrders"

	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("symbol", o.symbol)

	body, err := o.CallAPI(http.MethodGet, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var openOrdersResponse []QueryOrderResponse
	err = json.Unmarshal(body, &openOrdersResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return openOrdersResponse, nil
}
