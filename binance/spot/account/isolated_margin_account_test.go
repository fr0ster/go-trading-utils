package account_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	account_interface "github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/stretchr/testify/assert"
)

func TestIsolatedMarginAccountLimits_GetQuantityLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	// symbols := []string{"BTC", "ETH", "BNB", "USDT", "SUSHI", "CYBER"}
	symbols := []string{}
	spot := binance.NewClient(api_key, secret_key)
	account, err := spot_account.NewIsolatedMargin(spot, symbols)
	if err != nil {
		t.Errorf("Error creating account limits: %v", err)
		return
	}

	test := func(al account_interface.Accounts) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		freeAssets, err := al.GetAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, freeAssets)
	}

	test(account)
}

func TestIsolatedMarginAccountLimits_GetQuantityEmptyLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	symbols := []string{}
	spot := binance.NewClient(api_key, secret_key)
	account, err := spot_account.NewIsolatedMargin(spot, symbols)
	if err != nil {
		t.Errorf("Error creating account limits: %v", err)
		return
	}

	test := func(al account_interface.Accounts) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		freeAssets, err := al.GetAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, freeAssets)
	}

	test(account)
}
