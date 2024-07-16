package processor

import "github.com/fr0ster/go-trading-utils/types/depth/types"

func (pp *PairProcessor) NextPriceUp(price types.PriceType) types.PriceType {
	return types.PriceType(pp.RoundPrice(price * (1 + pp.GetDeltaPrice())))
}

func (pp *PairProcessor) NextPriceDown(price types.PriceType) types.PriceType {
	return types.PriceType(pp.RoundPrice(price * (1 - pp.GetDeltaPrice())))
}

func (pp *PairProcessor) NextQuantityUp(quantity types.QuantityType) types.QuantityType {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextQuantityDown(quantity types.QuantityType) types.QuantityType {
	return pp.RoundQuantity(quantity * (1 - pp.GetDeltaQuantity()))
}
