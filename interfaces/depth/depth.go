package depth

import (
	"github.com/adshao/go-binance/v2/common"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	Depths interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool) (err error)
		GetItem(price types.Price) *DepthItemType
		SetItem(value DepthItemType)
		UpdateAsk(ask common.PriceLevel, askLastUpdateID AskLastUpdateID) (err error)
		UpdateBid(bid common.PriceLevel, bidLastUpdateID BidLastUpdateID) (err error)
		GetMaxBids() *DepthItemType
		GetMaxAsks() *DepthItemType
		GetMaxBidQtyMaxAskQty() (maxBidNode *DepthItemType, maxAskNode *DepthItemType)
		GetMaxBidMinAsk() (maxBid *DepthItemType, minAsk *DepthItemType)
		GetBidQtyLocalMaxima() *btree.BTree
		GetAskQtyLocalMaxima() *btree.BTree
		Show()
	}
	DepthItemType struct {
		Price       types.Price
		AskQuantity types.Price
		BidQuantity types.Price
	}
	AskLastUpdateID int64
	BidLastUpdateID int64
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItemType).Price
}

func (i *DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItemType).Price
}
