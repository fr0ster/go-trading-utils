package depths

import (
	"sync"

	"github.com/google/btree"
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree int, symbol string) *Depths {
	return &Depths{
		symbol: symbol,
		degree: degree,

		tree:  btree.New(degree),
		mutex: &sync.Mutex{},

		countQuantity: 0,
		summaQuantity: 0,
		summaValue:    0,
	}
}

// Depths -
func (d *Depths) Symbol() string {
	return d.symbol
}

func (d *Depths) Degree() int {
	return d.degree
}
