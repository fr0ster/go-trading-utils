package balances_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/markets/balances"
	"github.com/stretchr/testify/assert"
)

func TestBalanceBTree(t *testing.T) {
	// Initialize the BalanceBTree
	balanceTree := balances.New(3, nil)

	// Create some sample balance items
	balanceItem1 := balances.BalanceItemType{
		Asset:  "BTC",
		Free:   1.0,
		Locked: 0.0,
	}
	balanceItem2 := balances.BalanceItemType{
		Asset:  "ETH",
		Free:   2.0,
		Locked: 0.5,
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
