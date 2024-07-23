package symbol_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spotInfo "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_info.New(spotInfo.InitCreator(2, client))
	symbol := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: "BTCUSDT"}).(*symbol_info.SpotSymbol)

	// Check if the struct implements the interface
	test := func(s *symbol_info.SpotSymbol) interface{} {
		return s.GetFilter("MAX_NUM_ALGO_ORDERS")
	}
	res := test(symbol)
	assert.NotNil(t, res)
}
