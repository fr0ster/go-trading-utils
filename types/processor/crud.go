package processor

import (
	"math"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func (pp *Processor) GetSymbol() string {
	return pp.symbol
}

func (pp *Processor) GetBaseSymbol() symbol_types.QuoteAsset {
	return pp.symbolInfo.GetBaseSymbol()
}

func (pp *Processor) GetTargetSymbol() symbol_types.BaseAsset {
	return pp.symbolInfo.GetTargetSymbol()
}

func (pp *Processor) GetSymbolInfo() *symbol_types.Symbol {
	return pp.symbolInfo
}

// Округлення ціни до StepSize знаків після коми
func (pp *Processor) GetStepSizeExp() (res int) {
	exp := int(math.Floor(math.Log10(float64(pp.symbolInfo.GetStepSize()))))
	res = -exp
	return
}

// Округлення ціни до TickSize знаків після коми
func (pp *Processor) GetTickSizeExp() (res int) {
	exp := int(math.Floor(math.Log10(float64(pp.symbolInfo.GetTickSize()))))
	res = -exp
	return
}

func (pp *Processor) GetNotional() items_types.ValueType {
	return pp.symbolInfo.GetNotional()
}

func (pp *Processor) GetMaxQty() items_types.QuantityType {
	return pp.symbolInfo.GetMaxQty()
}

func (pp *Processor) GetMinQty() items_types.QuantityType {
	return pp.symbolInfo.GetMinQty()
}

func (pp *Processor) GetMaxPrice() items_types.PriceType {
	return pp.symbolInfo.GetMaxPrice()
}

func (pp *Processor) GetMinPrice() items_types.PriceType {
	return pp.symbolInfo.GetMinPrice()
}

func (pp *Processor) GetCallbackRate() items_types.PricePercentType {
	if pp.getCallbackRate == nil {
		return 0
	}
	return pp.getCallbackRate()
}

func (pp *Processor) GetBaseBalance() items_types.ValueType {
	if pp.getBaseBalance == nil {
		return 0
	}
	return pp.getBaseBalance()
}
func (pp *Processor) GetTargetBalance() items_types.QuantityType {
	if pp.getTargetBalance == nil {
		return 0
	}
	return pp.getTargetBalance()

}
func (pp *Processor) GetFreeBalance() items_types.ValueType {
	if pp.getFreeBalance == nil {
		return 0
	}
	return pp.getFreeBalance()
}
func (pp *Processor) GetLockedBalance() items_types.ValueType {
	if pp.getLockedBalance == nil {
		return 0
	}
	return pp.getLockedBalance()
}
func (pp *Processor) GetCurrentPrice() items_types.PriceType {
	if pp.getCurrentPrice == nil {
		return 0
	}
	return pp.getCurrentPrice()
}
