package bookticker

import (
	items_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"
	"github.com/google/btree"
)

type (
	BookTicker interface {
		Lock()
		Unlock()
		Ascend(func(btree.Item) bool)
		Descend(func(btree.Item) bool)
		Get(string) *items_types.BookTicker
		Set(*items_types.BookTicker)
	}
)
