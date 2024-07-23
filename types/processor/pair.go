package processor

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *Processor) GetSymbol() string {
	return pp.symbol
}

// Округлення ціни до StepSize знаків після коми
func (pp *Processor) GetStepSizeExp() int {
	return int(pp.symbolInfo.StepSize)
}

// Округлення ціни до TickSize знаків після коми
func (pp *Processor) GetTickSizeExp() int {
	return int(pp.symbolInfo.tickSize)
}

func (pp *Processor) GetNotional() items_types.ValueType {
	return pp.symbolInfo.notional
}

func (pp *Processor) GetMaxQty() items_types.QuantityType {
	return pp.symbolInfo.maxQty
}

func (pp *Processor) GetMinQty() items_types.QuantityType {
	return pp.symbolInfo.minQty
}

func (pp *Processor) GetMaxPrice() items_types.PriceType {
	return pp.symbolInfo.maxPrice
}

func (pp *Processor) GetMinPrice() items_types.PriceType {
	return pp.symbolInfo.minPrice
}

func (pp *Processor) GetCallbackRate() items_types.PricePercentType {
	if pp.getCallbackRate == nil {
		return 0
	}
	return pp.getCallbackRate()
}
