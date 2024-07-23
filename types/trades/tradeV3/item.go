package tradeV3

import (
	"github.com/google/btree"
)

type (
	TradeV3 struct {
		ID              int64  `json:"id"`
		Symbol          string `json:"symbol"`
		OrderID         int64  `json:"orderId"`
		OrderListId     int64  `json:"orderListId"`
		Price           string `json:"price"`
		Quantity        string `json:"qty"`
		QuoteQuantity   string `json:"quoteQty"`
		Commission      string `json:"commission"`
		CommissionAsset string `json:"commissionAsset"`
		Time            int64  `json:"time"`
		IsBuyer         bool   `json:"isBuyer"`
		IsMaker         bool   `json:"isMaker"`
		IsBestMatch     bool   `json:"isBestMatch"`
		IsIsolated      bool   `json:"isIsolated"`
	}
)

func (tv3 *TradeV3) Less(than btree.Item) bool {
	return tv3.ID < than.(*TradeV3).ID
}

func (tv3 *TradeV3) Equal(than btree.Item) bool {
	return tv3.ID == than.(*TradeV3).ID
}
