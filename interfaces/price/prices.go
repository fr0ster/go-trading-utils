package kline

import "github.com/google/btree"

type (
	Prices interface {
		Lock()
		Unlock()
		Get(symbol string) btree.Item
		Set(value btree.Item)
	}
)
