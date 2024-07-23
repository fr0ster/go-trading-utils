package processor

import (
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/processor/symbol_info"
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

func (pp *Processor) GetSymbolInfo() *symbol_info_types.SymbolInfo {
	return pp.symbolInfo
}

// Округлення ціни до StepSize знаків після коми
func (pp *Processor) GetStepSizeExp() int {
	return int(pp.symbolInfo.StepSize)
}

// Округлення ціни до TickSize знаків після коми
func (pp *Processor) GetTickSizeExp() int {
	return int(pp.symbolInfo.GetTickSizeExp())
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
