package trades

import "github.com/google/btree"

type (
	AggTradeItem struct {
		AggTradeID       int64  `json:"a"`
		Price            string `json:"p"`
		Quantity         string `json:"q"`
		FirstTradeID     int64  `json:"f"`
		LastTradeID      int64  `json:"l"`
		Timestamp        int64  `json:"T"`
		IsBuyerMaker     bool   `json:"m"`
		IsBestPriceMatch bool   `json:"M"`
	}
	// WsTradeEvent - структура для зберігання даних про торги
	Trades interface {
		Lock()
		Unlock()
		// Init(apt_key, secret_key, symbolname string, UseTestnet bool) (err error)
		Ascend(iter func(btree.Item) bool)
		Descend(iter func(btree.Item) bool)
		Get(val *AggTradeItem) *AggTradeItem
		Set(val *AggTradeItem)
		Update(val *AggTradeItem)
	}
)

func (i *AggTradeItem) Less(than btree.Item) bool {
	return i.AggTradeID < than.(*AggTradeItem).AggTradeID
}
