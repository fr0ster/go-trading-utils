package spot_web_api

import "sync"

type (
	WebApi struct {
		apiKey    string
		apiSecret string
		symbol    string
		baseUrl   string
		mutex     *sync.Mutex
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
	return &WebApi{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		baseUrl:   GetAPIBaseUrl(useTestNet...),
	}
}
