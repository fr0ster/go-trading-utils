package trades

import "github.com/google/btree"

type (
	// WsTradeEvent - структура для зберігання даних про торги
	Trades interface {
		Lock()
		Unlock()
		// Init(apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error)
		Ascend(iter func(btree.Item) bool)
		Descend(iter func(btree.Item) bool)
		Get(id int64) btree.Item
		Set(val btree.Item)
		Update(val btree.Item)
	}
)
