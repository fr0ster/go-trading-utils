package kline

import "github.com/google/btree"

type (
	Prices interface {
		Lock()
		Unlock()
		Get(value btree.Item) btree.Item
		Set(value btree.Item)
	}
)
