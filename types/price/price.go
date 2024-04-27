package price

import (
	"sync"

	// prices_interface "github.com/fr0ster/go-trading-utils/interfaces/prices"
	"github.com/google/btree"
)

type (
	PriceChangeStats struct {
		tree   btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

func (d *PriceChangeStats) Get(value btree.Item) btree.Item {
	return d.tree.Get(value)
}

func (d *PriceChangeStats) Set(value btree.Item) {
	d.tree.ReplaceOrInsert(value)
}

func (d *PriceChangeStats) Lock() {
	d.mutex.Lock()
}

func (d *PriceChangeStats) Unlock() {
	d.mutex.Unlock()
}

// PriceChangeStats - B-дерево для зберігання Цінових змін
func New(degree int) *PriceChangeStats {
	return &PriceChangeStats{
		tree:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}
