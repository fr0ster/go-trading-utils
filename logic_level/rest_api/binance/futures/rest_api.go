package futures_rest_api

import (
	common "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/common"
	api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

type RestApi interface {
	// PlaceOrder() *order.Order
	// CancelOrder() *order.Order
	// QueryOrder() *order.Order
	// CancelReplaceOrder() *order.Order
	// QueryOpenOrders() *order.Order
	// QueryAllOrders() *order.Order
}

func New(apiKey, apiSecret string, symbol string, sign signature.Sign, useTestNet ...bool) RestApi {
	var (
		baseUrl api.ApiBaseUrl
	)
	if len(useTestNet) == 0 {
		useTestNet = append(useTestNet, false)
	}
	baseUrl = GetAPIBaseUrl(useTestNet...)
	return common.New(apiKey, apiSecret, baseUrl, symbol, sign)
}
