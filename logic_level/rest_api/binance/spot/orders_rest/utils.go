package orders_rest

import (
	"net/url"

	api "github.com/fr0ster/go-trading-utils/low_level/rest_api"
)

// Функція для отримання масиву всіх спотових ордерів
func (o *Orders) CallAPI(method string, params url.Values, endpoint string) (body []byte, err error) {
	return api.CallAPI(o.baseUrl, method, params, endpoint, o.sign)
}
