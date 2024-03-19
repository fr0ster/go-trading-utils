package account_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/account"
)

func TestAccountNew(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	account, err := account.New(client, 3)
	if err != nil || account == nil {
		t.Errorf("Error creating account: %v", err)
	}

	// Add assertions here to verify the account object
}

func TestAccountType_GetBalances(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	account, _ := account.New(client, 3)

	balances := account.GetBalances()

	if balances == nil {
		t.Errorf("Error creating balances")
	}

	// Add assertions here to verify the balances object
}

func TestAccountType_GetAccountInfo(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	account, _ := account.New(client, 3)

	accountInfo := account.GetAccountInfo()

	if accountInfo == nil {
		t.Errorf("Error creating accountInfo")
	}

	// Add assertions here to verify the accountInfo object
}
