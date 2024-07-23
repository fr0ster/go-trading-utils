package aggtrade

import (
	"github.com/google/btree"
)

type (
	AggTrade struct {
		AggTradeID       int64  `json:"a"`
		Price            string `json:"p"`
		Quantity         string `json:"q"`
		FirstTradeID     int64  `json:"f"`
		LastTradeID      int64  `json:"l"`
		Timestamp        int64  `json:"T"`
		IsBuyerMaker     bool   `json:"m"`
		IsBestPriceMatch bool   `json:"M"`
	}
)

func (a *AggTrade) Less(than btree.Item) bool {
	return a.AggTradeID < than.(*AggTrade).AggTradeID
}

func (a *AggTrade) Equal(than btree.Item) bool {
	return a.AggTradeID == than.(*AggTrade).AggTradeID
}
