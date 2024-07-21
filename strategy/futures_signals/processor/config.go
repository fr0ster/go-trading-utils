package processor

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (pp *PairProcessor) GetDeltaPrice() types.PriceType {
	return types.PriceType(pp.deltaPrice)
}

func (pp *PairProcessor) SetDeltaPrice(deltaPrice types.PriceType) {
	pp.deltaPrice = deltaPrice
}

func (pp *PairProcessor) GetDeltaQuantity() types.QuantityType {
	return types.QuantityType(pp.deltaQuantity)
}

func (pp *PairProcessor) GetLimitOnTransaction() (limit types.ValueType) {
	return types.ValueType(pp.limitOnTransaction) * pp.GetFreeBalance()
}

func (pp *PairProcessor) SetBounds(price types.PriceType) {
	pp.UpBound = price * (1 + types.PriceType(pp.UpBoundPercent))
	pp.LowBound = price * (1 - types.PriceType(pp.LowBoundPercent))
}

func (pp *PairProcessor) GetUpBound() types.PriceType {
	return pp.UpBound
}

func (pp *PairProcessor) GetLowBound() types.PriceType {
	return pp.LowBound
}
