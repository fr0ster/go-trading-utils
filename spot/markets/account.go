package markets

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
)

type AccountType struct {
	*binance.Account
	degree int
	sync.Mutex
}

func (account *AccountType) Lock() {
	account.Mutex.Lock()
}

func (account *AccountType) Unlock() {
	account.Mutex.Unlock()
}

func AccountNew(client *binance.Client, degree int) (*AccountType, error) {
	res, err := client.NewGetAccountService().Do(context.Background())
	return &AccountType{Account: res, degree: degree, Mutex: sync.Mutex{}}, err
}

func (account *AccountType) GetBalances() *BalanceBTree {
	balances := BalanceNew(account.degree, account.Account.Balances)
	return balances
}

func (account *AccountType) GetAccountInfo() *binance.Account {
	return account.Account
}
