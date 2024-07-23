package trade

import (
	"github.com/google/btree"
)

type (
	Trade struct {
		ID            int64  `json:"id"`
		Price         string `json:"price"`
		Quantity      string `json:"qty"`
		QuoteQuantity string `json:"quoteQty"`
		Time          int64  `json:"time"`
		IsBuyerMaker  bool   `json:"isBuyerMaker"`
		IsBestMatch   bool   `json:"isBestMatch"`
		IsIsolated    bool   `json:"isIsolated"`
	}
)

func (t *Trade) Less(than btree.Item) bool {
	return t.ID < than.(*Trade).ID
}

func (t *Trade) Equal(than btree.Item) bool {
	return t.ID == than.(*Trade).ID
}
