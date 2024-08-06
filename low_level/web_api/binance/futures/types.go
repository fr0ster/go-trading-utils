package spot_web_api

import (
	"sync"

	signature "github.com/fr0ster/go-trading-utils/low_level/common/signature"
)

type (
	WebApi struct {
		apiKey     string
		apiSecret  string
		symbol     string
		useTestNet bool
		baseUrl    string
		waHost     string
		waPath     string
		mutex      *sync.Mutex
		sign       signature.Sign
	}

	// Структура для параметрів запиту
	LogonParams struct {
		APIKey    string `json:"apiKey"`
		Signature string `json:"signature"`
		Timestamp int64  `json:"timestamp"`
	}

	LogonRequest struct {
		ID     string      `json:"id"`
		Method string      `json:"method"`
		Params LogonParams `json:"params"`
	}

	LogonResponse struct {
		APIKey           string `json:"apiKey"`
		AuthorizedSince  int64  `json:"authorizedSince"`
		ConnectedSince   int64  `json:"connectedSince"`
		ReturnRateLimits bool   `json:"returnRateLimits"`
		ServerTime       int64  `json:"serverTime"`
	}

	StatusRequest struct {
		ID     string `json:"id"`
		Method string `json:"method"`
	}

	LogoutRequest struct {
		ID     string `json:"id"`
		Method string `json:"method"`
	}
)

func (wa *WebApi) Lock() {
	wa.mutex.Lock()
}

func (wa *WebApi) Unlock() {
	wa.mutex.Unlock()
}

func New(apiKey, apiSecret, symbol string, sign signature.Sign, useTestNet ...bool) *WebApi {
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
	return &WebApi{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		symbol:     symbol,
		baseUrl:    GetAPIBaseUrl(useTestNet...),
		useTestNet: useTestNet[0],
		waHost:     waHost,
		waPath:     waPath,
		mutex:      &sync.Mutex{},
		sign:       sign,
	}
}
