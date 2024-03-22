package trades

import "github.com/google/btree"

type (
	Trades interface {
		Lock()
		Unlock()
		Ascend(iter func(btree.Item) bool)
		Descend(iter func(btree.Item) bool)
		Get(id int64) btree.Item
		Set(val btree.Item)
		Update(val btree.Item)
	}
)
