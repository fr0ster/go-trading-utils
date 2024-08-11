package order

import (
	"github.com/bitly/go-simplejson"

	rest_api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

type (
	Request struct {
		sign       signature.Sign
		apiBaseUrl rest_api.ApiBaseUrl
		endPoint   rest_api.EndPoint
		method     rest_api.HttpMethod
		params     *simplejson.Json
	}
)

func (rq *Request) Set(name string, value interface{}) *Request {
	rq.params.Set(name, value)
	return rq
}

func (rq *Request) Do() (response *simplejson.Json, err error) {
	response, err = rest_api.CallRestAPI(rq.apiBaseUrl, rq.method, rq.params, rq.endPoint, rq.sign)
	return
}

func New(apiKey, symbol string, method rest_api.HttpMethod, baseUrl rest_api.ApiBaseUrl, endPoint rest_api.EndPoint, params *simplejson.Json, sign signature.Sign) *Request {
	return &Request{
		sign:       sign,
		apiBaseUrl: baseUrl,
		endPoint:   endPoint,
		method:     method,
		params:     params,
	}
}
