package common_web_api

import (
	"sync"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

const (
	DepthAPILimit5    DepthAPILimit = 5
	DepthAPILimit10   DepthAPILimit = 10
	DepthAPILimit20   DepthAPILimit = 20
	DepthAPILimit50   DepthAPILimit = 50
	DepthAPILimit100  DepthAPILimit = 100
	DepthAPILimit500  DepthAPILimit = 500
	DepthAPILimit1000 DepthAPILimit = 1000
)

type (
	DepthAPILimit int
	WebApi        struct {
		apiKey    string
		apiSecret string
		symbol    string
		waHost    web_api.WsHost
		waPath    web_api.WsPath
		mutex     *sync.Mutex
		sign      signature.Sign
	}
	Result struct {
		APIKey           string `json:"apiKey"`
		AuthorizedSince  int64  `json:"authorizedSince"`
		ConnectedSince   int64  `json:"connectedSince"`
		ReturnRateLimits bool   `json:"returnRateLimits"`
		ServerTime       int64  `json:"serverTime"`
	}
)
