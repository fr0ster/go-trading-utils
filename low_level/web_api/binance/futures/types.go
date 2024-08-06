package spot_web_api

import "sync"

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
)

func (wa *WebApi) Lock() {
	wa.mutex.Lock()
}

func (wa *WebApi) Unlock() {
	wa.mutex.Unlock()
}

func New(apiKey, apiSecret, symbol string, useTestNet ...bool) *WebApi {
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
	}
}
