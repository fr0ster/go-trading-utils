package common_web_api

import (
	"sync"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

type (
	WebApi struct {
		apiKey    string
		apiSecret string
		symbol    string
		waHost    web_api.WsHost
		waPath    web_api.WsPath
		mutex     *sync.Mutex
		sign      signature.Sign
	}
	LogonResult struct {
		APIKey           string `json:"apiKey"`
		AuthorizedSince  int64  `json:"authorizedSince"`
		ConnectedSince   int64  `json:"connectedSince"`
		ReturnRateLimits bool   `json:"returnRateLimits"`
		ServerTime       int64  `json:"serverTime"`
	}
)
