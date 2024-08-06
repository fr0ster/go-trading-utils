package exchange_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/futures/exchange"
	"github.com/stretchr/testify/assert"
)

func TestGetExchangeInfo(t *testing.T) {
	exchangeInfo, err := exchange.New()
	assert.Nil(t, err)
	assert.NotNil(t, exchangeInfo)
}
