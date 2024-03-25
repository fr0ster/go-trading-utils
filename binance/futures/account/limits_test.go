package account_test

import (
	"context"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	account_interface "github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/fr0ster/go-trading-utils/utils"
)

func TestAccountLimits_GetQuantityLimits(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	symbols := []string{"BTC", "ETH", "BNB", "USDT", "SUSHI", "CYBER"}
	futures := futures.NewClient(api_key, secret_key)
	account := futures_account.NewAccountLimits(futures, symbols)

	test := func(al account_interface.AccountLimits) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		quantityLimits := al.GetQuantities()
		if quantityLimits == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
	}

	test(account)
}

func TestAccountLimits_GetQuantityEmptyLimits(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	symbols := []string{}
	futures := futures.NewClient(api_key, secret_key)
	account := futures_account.NewAccountLimits(futures, symbols)

	test := func(al account_interface.AccountLimits) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		quantityLimits := al.GetQuantities()
		if quantityLimits == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
	}

	test(account)
}

func TestAccountLimits_GetAssetUSDTLimits(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	symbols := []string{"USDT"}
	futures := futures.NewClient(api_key, secret_key)
	account := futures_account.NewAccountLimits(futures, symbols)

	test := func(al account_interface.AccountLimits) {
		if al == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		quantityLimits := al.GetQuantities()
		if quantityLimits == nil {
			t.Errorf("GetQuantityLimits returned an empty map")
		}
		accountResult, err := futures.NewGetAccountService().Do(context.Background())
		if err != nil {
			t.Errorf("GetAccountService returned an error")
		}
		quantityTest := accountResult.Assets
		test := 0.0
		for _, balance := range quantityTest {
			if balance.Asset == "USDT" {
				test = utils.ConvStrToFloat64(balance.AvailableBalance)
			}
		}
		quantity, _ := al.GetAsset("USDT")
		if quantity != test {
			t.Errorf("GetQuantity returned 0")
		}
	}

	test(account)
}
