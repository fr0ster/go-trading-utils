package items

import (
	"github.com/google/btree"

	depths_items "github.com/fr0ster/go-trading-utils/types/depths/items"
)

type (
	BookTicker struct {
		updateID    int64
		Symbol      string
		bidPrice    depths_items.PriceType
		bidQuantity depths_items.QuantityType
		askPrice    depths_items.PriceType
		askQuantity depths_items.QuantityType
	}
)

func (i *BookTicker) Less(than btree.Item) bool {
	return i.Symbol < than.(*BookTicker).Symbol
}

func (i *BookTicker) Equal(than btree.Item) bool {
	return i.Symbol == than.(*BookTicker).Symbol
}

func (btt *BookTicker) GetSymbol() string {
	return btt.Symbol
}

func (btt *BookTicker) GetBidPrice() depths_items.PriceType {
	return btt.bidPrice
}

func (btt *BookTicker) SetBidPrice(bidPrice depths_items.PriceType) {
	btt.bidPrice = bidPrice
}

func (btt *BookTicker) GetBidQuantity() depths_items.QuantityType {
	return btt.bidQuantity
}

func (btt *BookTicker) SetBidQuantity(bidQuantity depths_items.QuantityType) {
	btt.bidQuantity = bidQuantity
}

func (btt *BookTicker) GetBidValue() depths_items.ValueType {
	return depths_items.ValueType(btt.bidPrice) * depths_items.ValueType(btt.bidQuantity)
}

func (btt *BookTicker) GetAskPrice() depths_items.PriceType {
	return btt.askPrice
}

func (btt *BookTicker) SetAskPrice(askPrice depths_items.PriceType) {
	btt.askPrice = askPrice
}

func (btt *BookTicker) GetAskQuantity() depths_items.QuantityType {
	return btt.askQuantity
}

func (btt *BookTicker) SetAskQuantity(askQuantity depths_items.QuantityType) {
	btt.askQuantity = askQuantity
}

func (btt *BookTicker) GetAskValue() depths_items.ValueType {
	return depths_items.ValueType(btt.askPrice) * depths_items.ValueType(btt.askQuantity)
}

func (btt *BookTicker) GetUpdateID() int64 {
	return btt.updateID
}

func (btt *BookTicker) SetUpdateID(updateID int64) {
	btt.updateID = updateID
}

func New(
	symbol string,
	bidPrice depths_items.PriceType,
	bidQuantity depths_items.QuantityType,
	askPrice depths_items.PriceType,
	askQuantity depths_items.QuantityType,
	updateID ...int64) *BookTicker {
	var UpdateID int64
	if len(updateID) > 0 {
		UpdateID = updateID[0]
	}
	return &BookTicker{
		updateID:    UpdateID,
		Symbol:      symbol,
		bidPrice:    bidPrice,
		bidQuantity: bidQuantity,
		askPrice:    askPrice,
		askQuantity: askQuantity,
	}
}
