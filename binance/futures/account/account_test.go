package account_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	account_interface "github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/stretchr/testify/assert"
)

func TestAccount_GetQuantityLimits(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	symbols := []string{"BTC", "ETH", "BNB", "USDT", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account := futures_account.New(futures, 3, symbols)

	err := account.Update()
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
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account := futures_account.New(futures, 3, symbols)

	err := account.Update()
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

func TestAccount_GetAsset(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account := futures_account.New(futures, 3, symbols)

	err := account.Update()
	assert.Nil(t, err)

	assets := account.GetAssets()
	assert.NotEqual(t, 0, len(assets))
}

func TestAccount_GetPositions(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	symbols := []string{"USDT", "BTC", "ETH", "BNB", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret)
	account := futures_account.New(futures, 3, symbols)

	err := account.Update()
	assert.Nil(t, err)

	positions := account.GetPositions()
	assert.NotEqual(t, 0, len(positions))
}
