package account_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	account_interface "github.com/fr0ster/go-trading-utils/interfaces/account"
)

func TestAccountLimits_GetQuantityLimits(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	symbols := []string{"BTC", "ETH", "BNB", "USDT", "SUSHI", "CYBER"}
	spot := binance.NewClient(api_key, secret_key)
	account := spot_account.NewAccountLimits(spot, symbols)

	test := func(al account_interface.AccountLimits) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		quantityLimits := al.GetQuantityLimits()
		if quantityLimits == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
	}

	test(account)
}

func TestAccountLimits_GetQuantityEmptyLimits(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	symbols := []string{}
	spot := binance.NewClient(api_key, secret_key)
	account := spot_account.NewAccountLimits(spot, symbols)

	test := func(al account_interface.AccountLimits) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		quantityLimits := al.GetQuantityLimits()
		if quantityLimits == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
	}

	test(account)
}
