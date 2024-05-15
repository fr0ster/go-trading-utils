package grid

import (
	"sync"

	"github.com/google/btree"
)

type (
	Grid struct {
		tree btree.BTree
		mu   sync.Mutex
	}
)

func (g *Grid) Lock() {
	g.mu.Lock()
}

func (g *Grid) Unlock() {
	g.mu.Unlock()
}

func (g *Grid) Get(value btree.Item) btree.Item {
	return g.tree.Get(value)
}

func (g *Grid) Set(value btree.Item) {
	g.tree.ReplaceOrInsert(value)
}

func New() *Grid {
	return &Grid{
		tree: *btree.New(3),
		mu:   sync.Mutex{},
	}
}
