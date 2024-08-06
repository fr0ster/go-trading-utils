package exchange

import (
	"encoding/json"
	"log"
	"net/http"

	common "github.com/fr0ster/go-trading-utils/low_level/common/rest_api"
	api "github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/spot"
)

type (
	ExchangeInfo struct {
		Timezone        string        `json:"timezone"`
		ServerTime      int64         `json:"serverTime"`
		RateLimits      []RateLimit   `json:"rateLimits"`
		ExchangeFilters []interface{} `json:"exchangeFilters"`
		Symbols         []Symbol      `json:"symbols"`
	}
)

func New(useTestNet ...bool) (exchangeInfo *ExchangeInfo, err error) {
	baseUrl := api.GetAPIBaseUrl(useTestNet...)
	endpoint := "/api/v3/exchangeInfo"

	body, err := common.CallAPI(baseUrl, http.MethodGet, nil, endpoint, nil)
	if err != nil {
		return
	}

	// Десеріалізація JSON відповіді
	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
	}

	return
}

func (e *ExchangeInfo) Symbol(symbol string) *Symbol {
	for _, s := range e.Symbols {
		if s.Symbol == symbol {
			return &s
		}
	}
	return nil
}
