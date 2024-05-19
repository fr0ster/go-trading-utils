package grid

import (
	"sync"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"
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

func (g *Grid) Delete(value btree.Item) {
	g.tree.Delete(value)
}

func (g *Grid) Ascend(iter func(item btree.Item) bool) {
	g.tree.Ascend(iter)
}

func (g *Grid) Descend(iter func(item btree.Item) bool) {
	g.tree.Descend(iter)
}

func (g *Grid) Debug(pair, fl string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debugf("%s %s:", fl, pair)
		g.Descend(func(record btree.Item) bool {
			order := record.(*Record)
			logrus.Debugf(" Order %v on price %v OrderSide %v", order.GetOrderId(), order.GetPrice(), order.GetOrderSide())
			return true
		})
	}
}

func New() *Grid {
	return &Grid{
		tree: *btree.New(3),
		mu:   sync.Mutex{},
	}
}
