package processor

import (
	"math"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func (pp *Processor) GetDepths() *depth_types.Depths {
	return pp.depths
}

func (pp *Processor) SetDepths(depths *depth_types.Depths) {
	if pp.depths != nil {
		pp.depths.DepthEventStop()
		pp.depths = nil
	}
	pp.depths = depths
}

func (pp *Processor) GetOrders() *orders_types.Orders {
	return pp.orders
}

func (pp *Processor) SetOrders(orders *orders_types.Orders) {
	if pp.orders != nil {
		pp.orders.UserDataEventStop()
		pp.orders = nil
	}
	pp.orders = orders
}

func (pp *Processor) GetExchangeInfo() *exchange_types.ExchangeInfo {
	return pp.exchangeInfo
}

func (pp *Processor) SetExchangeInfo(exchangeInfo *exchange_types.ExchangeInfo) {
	pp.exchangeInfo = exchangeInfo
}

func (pp *Processor) GetSymbol() string {
	return pp.symbol
}

func (pp *Processor) GetBaseSymbol() symbol_types.QuoteAsset {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetBaseSymbol()
	return pp.symbolInfo.GetBaseSymbol()
}

func (pp *Processor) GetTargetSymbol() symbol_types.BaseAsset {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetTargetSymbol()
	return pp.symbolInfo.GetTargetSymbol()
}

func (pp *Processor) GetSymbolInfo() *symbol_types.Symbol {
	// return pp.exchangeInfo.GetSymbol(pp.symbol)
	return pp.symbolInfo
}

// Округлення ціни до StepSize знаків після коми
func (pp *Processor) GetStepSizeExp() (res int) {
	// exp := int(math.Floor(math.Log10(float64(pp.exchangeInfo.GetSymbol(pp.symbol).GetStepSize()))))
	exp := int(math.Floor(math.Log10(float64(pp.symbolInfo.GetStepSize()))))
	res = -exp
	return
}

// Округлення ціни до TickSize знаків після коми
func (pp *Processor) GetTickSizeExp() (res int) {
	// exp := int(math.Floor(math.Log10(float64(pp.exchangeInfo.GetSymbol(pp.symbol).GetTickSize()))))
	exp := int(math.Floor(math.Log10(float64(pp.symbolInfo.GetTickSize()))))
	res = -exp
	return
}

func (pp *Processor) GetNotional() items_types.ValueType {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetNotional()
	return pp.symbolInfo.GetNotional()
}

func (pp *Processor) GetMaxQty() items_types.QuantityType {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetMaxQty()
	return pp.symbolInfo.GetMaxQty()
}

func (pp *Processor) GetMinQty() items_types.QuantityType {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetMinQty()
	return pp.symbolInfo.GetMinQty()
}

func (pp *Processor) GetMaxPrice() items_types.PriceType {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetMaxPrice()
	return pp.symbolInfo.GetMaxPrice()
}

func (pp *Processor) GetMinPrice() items_types.PriceType {
	// return pp.exchangeInfo.GetSymbol(pp.symbol).GetMinPrice()
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
