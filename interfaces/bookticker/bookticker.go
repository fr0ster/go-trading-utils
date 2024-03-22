package bookticker

import (
	"github.com/google/btree"
)

type (
	BookTicker interface {
		Lock()
		Unlock()
		Ascend(func(btree.Item) bool)
		Descend(func(btree.Item) bool)
		Get(string) btree.Item
		Set(btree.Item)
	}
)
