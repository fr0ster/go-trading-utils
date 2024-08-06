package exchange_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/spot/exchange"
	"github.com/stretchr/testify/assert"
)

func TestGetExchangeInfo(t *testing.T) {
	exchangeInfo, err := exchange.New()
	assert.Nil(t, err)
	assert.NotNil(t, exchangeInfo)
	for _, symbol := range exchangeInfo.Symbols {
		assert.NotEmpty(t, symbol.NotionalFilter().MinNotional)
	}
}
