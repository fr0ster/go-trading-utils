package processor

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *Processor) GetDeltaPrice() items_types.PricePercentType {
	if pp.getDeltaPrice == nil {
		return 0
	}
	return pp.getDeltaPrice()
}

func (pp *Processor) GetDeltaQuantity() items_types.QuantityPercentType {
	if pp.getDeltaQuantity == nil {
		return 0
	}
	return pp.getDeltaQuantity()
}

func (pp *Processor) GetLimitOnPosition() (limit items_types.ValueType) {
	if pp.getLimitOnPosition == nil {
		return 0
	}
	if pp.GetFreeBalance() > pp.getLimitOnPosition() {
		return pp.getLimitOnPosition()
	} else {
		return pp.GetFreeBalance()
	}
}

func (pp *Processor) GetLimitOnTransaction() (limit items_types.ValueType) {
	if pp.getLimitOnTransaction == nil {
		return 0
	}
	return items_types.ValueType(pp.getLimitOnTransaction()/100) * pp.GetLimitOnPosition()
}

func (pp *Processor) GetUpAndLowBound() items_types.PricePercentType {
	if pp.getUpAndLowBound == nil {
		return 0
	}
	bound := pp.getUpAndLowBound()
	if bound == 0 {
		return 100 / items_types.PricePercentType(pp.GetLeverage())
	} else {
		return bound
	}
}

func (pp *Processor) GetUpBound(price items_types.PriceType) items_types.PriceType {
	if pp.getUpAndLowBound == nil {
		return price
	}
	return price * (1 + items_types.PriceType(pp.getUpAndLowBound()/100))
}

func (pp *Processor) GetLowBound(price items_types.PriceType) items_types.PriceType {
	if pp.getUpAndLowBound == nil {
		return price
	}
	return price * (1 - items_types.PriceType(pp.getUpAndLowBound()/100))
}
