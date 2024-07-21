package depths

import (
	"sync"

	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

const (
	UP   UpOrDown = true
	DOWN UpOrDown = false
)

type (
	UpOrDown bool
	Depths   struct {
		symbol string
		degree int

		tree  *btree.BTree
		mutex *sync.Mutex

		countQuantity int
		summaQuantity items_types.QuantityType
		summaValue    items_types.ValueType
	}
)
