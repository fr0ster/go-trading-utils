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
	symbols := []string{"BTC", "ETH", "BNB", "USDT", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, symbols)
	assert.Nil(t, err)

	err = account.Update()
	assert.Nil(t, err)

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

func TestAccount_GetQuantityEmptyLimits(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, symbols)
	assert.Nil(t, err)

	err = account.Update()
	assert.Nil(t, err)

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

func TestAccount_GetAssets(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, symbols)
	assert.Nil(t, err)

	err = account.Update()
	assert.Nil(t, err)

	assets := account.GetAssets()
	assert.NotEqual(t, 0, len(assets))
}

func TestAccount_GetPositions(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret)
	account, err := futures_account.New(futures, 3, symbols)
	assert.Nil(t, err)

	err = account.Update()
	assert.Nil(t, err)

	positions := account.GetPositions()
	assert.NotEqual(t, 0, len(positions))
}

func TestAccount_GetAsset(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHIUSDT"}
	futures := futures.NewClient(api_key, secret_key)
	account, err := futures_account.New(futures, 3, symbols)
	assert.Nil(t, err)

	_, err = account.GetAsset("SUSHIUSDT")
	assert.Nil(t, err)
}
