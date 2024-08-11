package futures_web_api

import (
	common "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common"
	request "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/request"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

type WebApi interface {
	PlaceOrder() *request.Request
	CancelOrder() *request.Request
	QueryOrder() *request.Request
	CancelReplaceOrder() *request.Request
	ListOfSubscriptions() *request.Request
	Logon() (result *common.Result, err error)
	Logout() (result *common.Result, err error)
	Status() (result *common.Result, err error)
}

func New(apiKey, apiSecret, symbol string, sign signature.Sign, useTestNet ...bool) WebApi {
	var (
		waHost string
		waPath string
	)
	if len(useTestNet) == 0 {
		useTestNet = append(useTestNet, false)
	}
	if useTestNet[0] {
		waHost = "testnet.binancefuture.com"
		waPath = "/ws-fapi/v1"
	} else {
		waHost = "ws-fapi.binance.com"
		waPath = "/ws-fapi/v1"
	}
	return common.New(apiKey, apiSecret, web_api.WsHost(waHost), web_api.WsPath(waPath), symbol, sign)
}
