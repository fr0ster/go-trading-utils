package grid

import (
	"sync"

	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"
)

type (
	Grid struct {
		tree            btree.BTree
		countSellOrders int
		countBuyOrders  int
		mu              sync.Mutex
	}
)

func (g *Grid) Lock() {
	g.mu.Lock()
}

func (g *Grid) Unlock() {
	g.mu.Unlock()
}

func (g *Grid) GetCountSellOrders() int {
	return g.countSellOrders
}

func (g *Grid) GetCountBuyOrders() int {
	return g.countBuyOrders
}

func (g *Grid) Get(value btree.Item) btree.Item {
	return g.tree.Get(value)
}

func (g *Grid) Set(value btree.Item) {
	record := value.(*Record)
	if record.GetOrderSide() == types.SideTypeSell {
		g.countSellOrders++
	} else if record.GetOrderSide() == types.SideTypeBuy {
		g.countBuyOrders++
	}
	g.tree.ReplaceOrInsert(value)
}

func (g *Grid) Delete(value btree.Item) {
	g.tree.Delete(value)
	record := value.(*Record)
	if record.GetOrderSide() == types.SideTypeSell {
		g.countSellOrders--
	} else if record.GetOrderSide() == types.SideTypeBuy {
		g.countBuyOrders--
	}
}

func (g *Grid) Ascend(iter func(item btree.Item) bool) {
	g.tree.Ascend(iter)
}

func (g *Grid) Descend(iter func(item btree.Item) bool) {
	g.tree.Descend(iter)
}

func (g *Grid) CancelSellOrder() {
	g.Descend(func(record btree.Item) bool {
		order := record.(*Record)
		if order.GetOrderSide() == types.SideTypeBuy {
			g.Delete(order)
		} else if order.GetOrderSide() == types.SideTypeNone {
			order.SetUpPrice(0)
		}
		return true
	})
}

func (g *Grid) CancelBuyOrder() {
	g.Descend(func(record btree.Item) bool {
		order := record.(*Record)
		if order.GetOrderSide() == types.SideTypeSell {
			g.Delete(order)
		} else if order.GetOrderSide() == types.SideTypeNone {
			order.SetUpPrice(0)
		}
		return true
	})
}

func (g *Grid) Debug(pair, id, fl string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debugf("%s %s %s:", fl, id, pair)
		g.Descend(func(record btree.Item) bool {
			order := record.(*Record)
			logrus.Debugf(" Order %v on price %v with quantity %v OrderSide %v", order.GetOrderId(), order.GetPrice(), order.GetQuantity(), order.GetOrderSide())
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
