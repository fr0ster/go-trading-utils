package processor

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *Processor) GetDeltaPrice() items_types.PriceType {
	if pp.getDeltaPrice == nil {
		return 0
	}
	return pp.getDeltaPrice()
}

func (pp *Processor) GetDeltaQuantity() items_types.QuantityType {
	if pp.getDeltaQuantity == nil {
		return 0
	}
	return pp.getDeltaQuantity()
}

func (pp *Processor) GetLimitOnTransaction() (limit items_types.ValueType) {
	if pp.getLimitOnTransaction == nil {
		return 0
	}
	return pp.getLimitOnTransaction() * pp.GetFreeBalance()
}

func (pp *Processor) GetUpBound(price items_types.PriceType) items_types.PriceType {
	if pp.getUpAndLowBound == nil {
		return price
	}
	return price * (1 + items_types.PriceType(pp.getUpAndLowBound()))
}

func (pp *Processor) GetLowBound(price items_types.PriceType) items_types.PriceType {
	if pp.getUpAndLowBound == nil {
		return price
	}
	return price * (1 - items_types.PriceType(pp.getUpAndLowBound()))
}
