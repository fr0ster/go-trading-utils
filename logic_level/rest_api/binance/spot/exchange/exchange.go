package exchange

import (
	"encoding/json"
	"log"
	"net/http"

	api "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/spot"
	common "github.com/fr0ster/turbo-restler/rest_api"
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

	body, err := common.CallRestAPI(baseUrl, http.MethodGet, nil, endpoint, nil)
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
