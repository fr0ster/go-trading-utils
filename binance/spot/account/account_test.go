package account_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	account_interface "github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/stretchr/testify/assert"
)

const (
	API_KEY      = "API_KEY"
	SECRET_KEY   = "SECRET_KEY"
	USE_TEST_NET = false
)

func TestAccountLimits_GetQuantityLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	symbols := []string{"BTC", "ETH", "BNB", "USDT", "SUSHI", "CYBER"}
	spot := binance.NewClient(api_key, secret_key)
	account, err := spot_account.New(spot, symbols)
	if err != nil {
		t.Errorf("Error creating account limits: %v", err)
		return
	}

	test := func(al account_interface.Accounts) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		free, err := al.GetFreeAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, free)
		locked, err := al.GetLockedAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, locked)
		total, err := al.GetTotalAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, total)
	}

	test(account)
}

func TestAccountLimits_GetQuantityEmptyLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	symbols := []string{}
	spot := binance.NewClient(api_key, secret_key)
	account, err := spot_account.New(spot, symbols)
	if err != nil {
		t.Errorf("Error creating account limits: %v", err)
		return
	}

	test := func(al account_interface.Accounts) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		free, err := al.GetFreeAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, free)
		locked, err := al.GetLockedAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, locked)
		total, err := al.GetTotalAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, total)
	}

	test(account)
}
