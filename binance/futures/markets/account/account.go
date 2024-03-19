package account

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/balances"
)

type AccountType struct {
	*futures.Account
	degree int
	sync.Mutex
}

func (account *AccountType) Lock() {
	account.Mutex.Lock()
}

func (account *AccountType) Unlock() {
	account.Mutex.Unlock()
}

func New(client *futures.Client, degree int) (*AccountType, error) {
	res, err := client.NewGetAccountService().Do(context.Background())
	return &AccountType{Account: res, degree: degree, Mutex: sync.Mutex{}}, err
}

func (account *AccountType) GetBalances() *balances.BalanceBTree {
	balances := balances.New(account.degree, account.Account.Assets)
	return balances
}

func (account *AccountType) GetAccountInfo() *futures.Account {
	return account.Account
}
