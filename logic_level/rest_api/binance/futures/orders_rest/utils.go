package orders_rest

import (
	"github.com/bitly/go-simplejson"
	api "github.com/fr0ster/turbo-restler/rest_api"
)

// Функція для отримання масиву всіх спотових ордерів
func (o *Orders) CallAPI(method api.HttpMethod, params *simplejson.Json, endpoint api.EndPoint) (body []byte, err error) {
	return api.CallRestAPI(o.baseUrl, method, params, endpoint, o.sign)
}
