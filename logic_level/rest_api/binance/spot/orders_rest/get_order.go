package orders_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
)

// Функція для отримання одного спотового ордера по номеру
func (o *Orders) GetOrder(orderID int64) (*QueryOrderResponse, error) {
	endpoint := "/api/v3/order"

	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("symbol", o.symbol)
	params.Set("orderId", strconv.FormatInt(orderID, 10))

	body, err := o.CallAPI(http.MethodGet, params, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// Десеріалізація JSON відповіді
	var orderResponse QueryOrderResponse
	err = json.Unmarshal(body, &orderResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &orderResponse, nil
}
