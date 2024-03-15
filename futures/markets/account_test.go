package markets_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-binance-utils/futures/markets"
	"github.com/stretchr/testify/assert"
)

func TestAccountNew(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	client := futures.NewClient(api_key, secret_key)
	account, err := markets.AccountNew(client)
	if err != nil || account == nil {
		t.Errorf("Error creating account: %v", err)
	}

	// Add assertions here to verify the account object
}

func TestAccountType_GetAccountInfo(t *testing.T) {
	api_key := os.Getenv("FUTURE_TEST_BINANCE_API_KEY")
	secret_key := os.Getenv("FUTURE_TEST_BINANCE_SECRET_KEY")
	futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	account, err := markets.AccountNew(client)
	assert.NoError(t, err)

	accountInfo := account.GetAccountInfo()

	if accountInfo == nil {
		t.Errorf("Error creating accountInfo")
	}

	// Add assertions here to verify the accountInfo object
}
