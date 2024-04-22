package symbol_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futuresInfo "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/stretchr/testify/assert"
)

const degree = 3

func TestInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	exchangeInfo := exchange_info.New()
	err := futuresInfo.Init(exchangeInfo, degree, client)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	symbol := exchangeInfo.GetSymbol(&symbol_info.FuturesSymbol{Symbol: "BTCUSDT"}).(*symbol_info.FuturesSymbol)

	// Check if the struct implements the interface
	test := func(s *symbol_info.FuturesSymbol) interface{} {
		return s.GetFilter("MAX_NUM_ALGO_ORDERS")
	}
	res := test(symbol)
	assert.NotNil(t, res)
}
