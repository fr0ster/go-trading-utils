package order

import (
	"fmt"

	rest_api "github.com/fr0ster/turbo-restler/rest_api"

	"github.com/bitly/go-simplejson"
	"github.com/fr0ster/turbo-restler/utils/signature"
)

type (
	Request struct {
		sign       signature.Sign
		apiBAseUrl rest_api.ApiBaseUrl
		endPoint   rest_api.EndPoint
		method     rest_api.HttpMethod
		params     *simplejson.Json
	}
)

func (rq *Request) Set(name string, value interface{}) *Request {
	rq.params.Set(name, value)
	return rq
}

func (rq *Request) Do() (order *simplejson.Json, err error) {
	response, err := rest_api.CallRestAPI(rq.apiBAseUrl, rq.method, rq.params, rq.endPoint, rq.sign)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	order, err = simplejson.NewJson(response)
	return
}

func New(apiKey, symbol string, method rest_api.HttpMethod, baseUrl rest_api.ApiBaseUrl, endPoint rest_api.EndPoint, sign signature.Sign) *Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &Request{
		sign:       sign,
		apiBAseUrl: baseUrl,
		endPoint:   endPoint,
		method:     method,
		params:     simpleJson,
	}
}
