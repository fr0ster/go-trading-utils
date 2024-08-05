package orders_rest

import (
	"sync"

	api "github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/futures"
)

func New(apiKey, apiSecret, symbol string, useTestNet ...bool) *Orders {
	return &Orders{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		baseUrl:   api.GetAPIBaseUrl(useTestNet...),
		mutex:     &sync.Mutex{},
	}
}

func (o *Orders) Lock() {
	o.mutex.Lock()
}

func (o *Orders) Unlock() {
	o.mutex.Unlock()
}
