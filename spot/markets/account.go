package markets

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
)

type AccountType struct {
	*binance.Account
	sync.Mutex
}

func AccountNew(client *binance.Client) (*AccountType, error) {
	res, err := client.NewGetAccountService().Do(context.Background())
	return &AccountType{Account: res, Mutex: sync.Mutex{}}, err
}

func (account *AccountType) GetBalances() *BalanceBTree {
	balances := BalanceNew(3)
	balances.Init(account.Account.Balances)
	return balances
}

func (account *AccountType) GetAccountInfo() *binance.Account {
	return account.Account
}
