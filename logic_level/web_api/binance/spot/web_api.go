package spot_web_api

import (
	common "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common"
	request "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/request"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	"github.com/fr0ster/turbo-restler/web_api"
)

type WebApi interface {
	PlaceRequest() *request.Request
	CancelRequest() *request.Request
	QueryRequest() *request.Request
	CancelReplaceRequest() *request.Request
	QueryOpenRequests() *request.Request
	QueryAllRequests() *request.Request
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
		waHost = "testnet.binance.vision"
		waPath = "/ws-api/v3"
	} else {
		waHost = "ws-api.binance.com:443"
		waPath = "/ws-api/v3"
	}
	return common.New(apiKey, apiSecret, web_api.WsHost(waHost), web_api.WsPath(waPath), symbol, sign)
}
