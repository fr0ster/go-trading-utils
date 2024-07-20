package processor

import (
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (pp *PairProcessor) NextPriceUp(price ...items.PriceType) items.PriceType {
	return pp.RoundPrice(pp.GetDepth().NextPriceUp(float64(pp.GetDeltaPrice()), price...))
}

func (pp *PairProcessor) NextPriceDown(price ...items.PriceType) items.PriceType {
	return pp.RoundPrice(pp.GetDepth().NextPriceDown(float64(pp.GetDeltaPrice()), price...))
}

func (pp *PairProcessor) NextQuantityUp(quantity items.QuantityType) items.QuantityType {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextQuantityDown(quantity items.QuantityType) items.QuantityType {
	return pp.RoundQuantity(quantity * (1 - pp.GetDeltaQuantity()))
}
