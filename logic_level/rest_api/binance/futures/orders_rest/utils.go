package orders_rest

import (
	"github.com/bitly/go-simplejson"
	api "github.com/fr0ster/turbo-restler/rest_api"
)

// Функція для отримання масиву всіх спотових ордерів
func (o *Orders) CallAPI(method string, params *simplejson.Json, endpoint string) (body []byte, err error) {
	return api.CallRestAPI(o.baseUrl, method, params, endpoint, o.sign)
}
