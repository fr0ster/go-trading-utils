package types

import "github.com/google/btree"

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

func (i AggTrade) Less(than btree.Item) bool {
	return i.AggTradeID < than.(AggTrade).AggTradeID
}

func (i AggTrade) Equal(than btree.Item) bool {
	return i.AggTradeID == than.(AggTrade).AggTradeID
}

func (i Trade) Less(than btree.Item) bool {
	return i.ID < than.(Trade).ID
}

func (i Trade) Equal(than btree.Item) bool {
	return i.ID == than.(Trade).ID
}

func (i TradeV3) Less(than btree.Item) bool {
	return i.ID < than.(TradeV3).ID
}

func (i TradeV3) Equal(than btree.Item) bool {
	return i.ID == than.(TradeV3).ID
}
