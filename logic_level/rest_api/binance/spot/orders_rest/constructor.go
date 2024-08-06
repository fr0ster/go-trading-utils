package orders_rest

import (
	"sync"

	api "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/spot"
	signature "github.com/fr0ster/go-trading-utils/low_level/utils/signature"
)

func New(apiKey, apiSecret, symbol string, sign signature.Sign, useTestNet ...bool) *Orders {
	return &Orders{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		baseUrl:   api.GetAPIBaseUrl(useTestNet...),
		mutex:     &sync.Mutex{},
		sign:      sign,
	}
}

func (o *Orders) Lock() {
	o.mutex.Lock()
}

func (o *Orders) Unlock() {
	o.mutex.Unlock()
}
