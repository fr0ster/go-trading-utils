package depth

import (
	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	Depths struct {
		client *binance.Client
		depth  btree.BTree
	}
)
