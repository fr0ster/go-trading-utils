package markets_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/futures/markets"
	"github.com/stretchr/testify/assert"
)

func TestBalanceBTree(t *testing.T) {
	// Initialize the BalanceBTree
	balanceTree := markets.BalanceNew(3, nil)

	// Create some sample balance items
	balanceItem1 := markets.BalanceItemType{
		Asset:              "BTC",
		Balance:            1.0,
		CrossWalletBalance: 0.5,
		ChangeBalance:      0.0,
	}
	balanceItem2 := markets.BalanceItemType{
		Asset:              "ETH",
		Balance:            2.0,
		CrossWalletBalance: 1.0,
		ChangeBalance:      0.0,
	}

	// Set the balance items in the tree
	balanceTree.SetItem(balanceItem1)
	balanceTree.SetItem(balanceItem2)

	// Get the balance item by asset
	result, err := balanceTree.GetItem("BTC")
	assert.NoError(t, err)
	assert.Equal(t, balanceItem1, result)

	// Show the balances tree
	balanceTree.Show()
}
