package kline

import "github.com/google/btree"

type (
	Prices interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool)
		Get(symbol string) btree.Item
		Set(value btree.Item)
	}
)
