package processor

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *PairProcessor) GetDeltaPrice() items_types.PriceType {
	return items_types.PriceType(pp.deltaPrice)
}

func (pp *PairProcessor) SetDeltaPrice(deltaPrice items_types.PriceType) {
	pp.deltaPrice = deltaPrice
}

func (pp *PairProcessor) GetDeltaQuantity() items_types.QuantityType {
	return items_types.QuantityType(pp.deltaQuantity)
}

func (pp *PairProcessor) GetLimitOnTransaction() (limit items_types.ValueType) {
	return items_types.ValueType(pp.limitOnTransaction) * pp.GetFreeBalance()
}

// func (pp *PairProcessor) SetBounds(price items_types.PriceType) {
// 	pp.UpBound = price * (1 + items_types.PriceType(pp.UpBoundPercent))
// 	pp.LowBound = price * (1 - items_types.PriceType(pp.LowBoundPercent))
// }

// func (pp *PairProcessor) GetUpBound() items_types.PriceType {
// 	return pp.UpBound
// }

// func (pp *PairProcessor) GetLowBound() items_types.PriceType {
// 	return pp.LowBound
// }

func (pp *PairProcessor) GetUpBound(price items_types.PriceType) items_types.PriceType {
	return price * (1 + items_types.PriceType(pp.UpBoundPercent))
}

func (pp *PairProcessor) GetLowBound(price items_types.PriceType) items_types.PriceType {
	return price * (1 - items_types.PriceType(pp.LowBoundPercent))
}
