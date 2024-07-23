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

func (pp *Processor) GetOrders() *orders_types.Orders {
	return pp.orders
}

func (pp *Processor) GetExchangeInfo() *exchange_types.ExchangeInfo {
	return pp.exchangeInfo
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

func (pp *Processor) GetSymbolInfo() *symbol_types.SymbolInfo {
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
