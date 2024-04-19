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
		balance        *btree.BTree
		mu             *sync.Mutex
		assetsName     map[string]bool
		assetsRestrict []string
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

func (b *Balances) Insert(balance *Balance) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.balance.ReplaceOrInsert(balance)
}

func (b *Balances) Delete(balance *Balance) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.balance.Delete(balance)
}

func (b *Balances) Get(asset string) *Balance {
	b.mu.Lock()
	defer b.mu.Unlock()
	item := b.balance.Get(&Balance{Asset: asset})
	if item == nil {
		return nil
	}
	return item.(*Balance)
}

func (b *Balances) Update(balance *Balance) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.balance.ReplaceOrInsert(balance)
}

func New(client *futures.Client, assets []string) (*Balances, error) {
	bl := &Balances{
		balance:        btree.New(2),
		mu:             &sync.Mutex{},
		assetsName:     make(map[string]bool),
		assetsRestrict: assets,
	}
	for _, asset := range bl.assetsRestrict {
		bl.assetsName[asset] = true
	}
	balances, err := client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	for _, balance := range balances {
		if _, exists := bl.assetsName[balance.Asset]; exists || len(bl.assetsName) == 0 {
			val := Balance(*balance)
			bl.balance.ReplaceOrInsert(&val)
		}
	}
	return bl, nil
}
