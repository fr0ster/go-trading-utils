package processor

import (
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
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetBaseSymbol()
}

func (pp *Processor) GetTargetSymbol() symbol_types.BaseAsset {
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetTargetSymbol()
}

func (pp *Processor) GetSymbolInfo() *symbol_types.Symbol {
	return pp.exchangeInfo.GetSymbol(pp.symbol)
}

// Округлення ціни до StepSize знаків після коми
func (pp *Processor) GetStepSizeExp() int {
	return int(pp.exchangeInfo.GetSymbol(pp.symbol).GetStepSize())
}

// Округлення ціни до TickSize знаків після коми
func (pp *Processor) GetTickSizeExp() int {
	return int(pp.exchangeInfo.GetSymbol(pp.symbol).GetTickSizeExp())
}

func (pp *Processor) GetNotional() items_types.ValueType {
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetNotional()
}

func (pp *Processor) GetMaxQty() items_types.QuantityType {
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetMaxQty()
}

func (pp *Processor) GetMinQty() items_types.QuantityType {
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetMinQty()
}

func (pp *Processor) GetMaxPrice() items_types.PriceType {
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetMaxPrice()
}

func (pp *Processor) GetMinPrice() items_types.PriceType {
	return pp.exchangeInfo.GetSymbol(pp.symbol).GetMinPrice()
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
func (pp *Processor) GetTargetBalance() items_types.ValueType {
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

// getBaseBalance func(*Processor) func() items_types.ValueType,
func (pp *Processor) SetGetBaseBalance(f func() items_types.ValueType) {
	if f != nil {
		pp.getBaseBalance = f
	}
}
