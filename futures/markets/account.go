package markets

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
)

type AccountType struct {
	*futures.Account
	sync.Mutex
}

func (account *AccountType) Lock() {
	account.Mutex.Lock()
}

func (account *AccountType) Unlock() {
	account.Mutex.Unlock()
}

func AccountNew(client *futures.Client) (*AccountType, error) {
	res, err := client.NewGetAccountService().Do(context.Background())
	return &AccountType{Account: res, Mutex: sync.Mutex{}}, err
}

func (account *AccountType) GetAccountInfo() *futures.Account {
	return account.Account
}
