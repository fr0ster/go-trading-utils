package balances

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
)

type (
	Balance  futures.Balance
	Balances struct {
		balance *btree.BTree
		mu      *sync.Mutex
	}
)

func (b *Balance) Less(item btree.Item) bool {
	return b.Asset < item.(*Balance).Asset
}

func (b *Balance) Equal(item btree.Item) bool {
	return b.Asset == item.(*Balance).Asset
}

func (b *Balances) Ascend(f func(item btree.Item) bool) {
	b.balance.Ascend(func(i btree.Item) bool {
		return f(i)
	})
}

func (b *Balances) Descend(f func(item btree.Item) bool) {
	b.balance.Descend(func(i btree.Item) bool {
		return f(i)
	})
}

func New(client *futures.Client) (*Balances, error) {
	bl := &Balances{
		balance: btree.New(2),
		mu:      &sync.Mutex{},
	}
	balances, err := client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	for _, balance := range balances {
		val := Balance(*balance)
		bl.balance.ReplaceOrInsert(&val)
	}
	return bl, nil
}
