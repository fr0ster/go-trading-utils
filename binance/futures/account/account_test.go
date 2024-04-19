package account_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	account_interface "github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/stretchr/testify/assert"
)

const (
	API_KEY      = "FUTURE_TEST_BINANCE_API_KEY"
	SECRET_KEY   = "FUTURE_TEST_BINANCE_SECRET_KEY"
	USE_TEST_NET = true
)

func TestAccount_GetQuantityLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	assets := []string{"BTC", "ETH", "BNB", "USDT"}
	symbols := []string{"SUSHIUSDT", "CYBERUSDT"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, assets, symbols)
	assert.Nil(t, err)

	test := func(al account_interface.Accounts) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		freeAssets, err := al.GetFreeAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, freeAssets)
	}

	test(account)
}

func TestAccount_GetQuantityEmptyLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	assets := []string{"BTC", "ETH", "BNB", "USDT"}
	symbols := []string{"SUSHIUSDT", "CYBERUSDT"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, assets, symbols)
	assert.Nil(t, err)

	test := func(al account_interface.Accounts) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		freeAssets, err := al.GetFreeAsset("USDT")
		assert.Nil(t, err)
		assert.NotEqual(t, 0, freeAssets)
	}

	test(account)
}

func TestAccount_GetAsset(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	assets := []string{"BTC", "ETH", "BNB", "USDT"}
	symbols := []string{"SUSHIUSDT", "CYBERUSDT"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, assets, symbols)
	assert.Nil(t, err)

	_, err = account.GetFreeAsset("SUSHIUSDT")
	assert.Nil(t, err)
}
