package balances_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"

	"github.com/fr0ster/go-trading-utils/binance/futures/balances"
)

const (
	API_KEY      = "FUTURE_TEST_BINANCE_API_KEY"
	SECRET_KEY   = "FUTURE_TEST_BINANCE_SECRET_KEY"
	USE_TEST_NET = true
)

var (
	b1 = &balances.Balance{
		AccountAlias:       "accountAlias1",
		Asset:              "BTC",
		Balance:            "1",
		CrossWalletBalance: "1",
		CrossUnPnl:         "1",
		AvailableBalance:   "1",
		MaxWithdrawAmount:  "1",
	}
	b2 = &balances.Balance{
		AccountAlias:       "accountAlias1",
		Asset:              "BTC",
		Balance:            "1",
		CrossWalletBalance: "1",
		CrossUnPnl:         "1",
		AvailableBalance:   "1",
		MaxWithdrawAmount:  "1",
	}
	b3 = &balances.Balance{
		AccountAlias:       "accountAlias2",
		Asset:              "ETH",
		Balance:            "1",
		CrossWalletBalance: "1",
		CrossUnPnl:         "1",
		AvailableBalance:   "1",
		MaxWithdrawAmount:  "1",
	}
	assets = []string{"BTC", "ETH"}
)

func createBalances(assets []string) (*balances.Balances, error) {

	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	futures.UseTestnet = USE_TEST_NET
	futures := futures.NewClient(api_key, secret_key)

	b, err := balances.New(futures, assets)
	return b, err
}

func TestNew(t *testing.T) {
	b, err := createBalances(assets)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}

func TestBalance_Less(t *testing.T) {
	assert.True(t, b1.Less(b3))
	assert.False(t, b2.Less(b1))
}

func TestBalance_Equal(t *testing.T) {
	assert.True(t, b1.Equal(b2))
	assert.False(t, b1.Equal(b3))
}

func TestBalances_Ascend(t *testing.T) {
	b, err := createBalances(assets)
	assert.NoError(t, err)
	b.Insert(b1)
	b.Insert(b2)
	b.Insert(b3)

	var result []string
	b.Ascend(func(item btree.Item) bool {
		balance := item.(*balances.Balance)
		result = append(result, balance.Asset)
		return true
	})

	expected := []string{"BTC", "ETH"}
	assert.Equal(t, expected, result)
}

func TestBalances_Descend(t *testing.T) {

	b, err := createBalances(assets)
	assert.NoError(t, err)
	b.Insert(b1)
	b.Insert(b2)
	b.Insert(b3)

	var result []string
	b.Descend(func(item btree.Item) bool {
		balance := item.(*balances.Balance)
		result = append(result, balance.Asset)
		return true
	})

	expected := []string{"ETH", "BTC"}
	assert.Equal(t, expected, result)
}
