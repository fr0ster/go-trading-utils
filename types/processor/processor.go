package processor

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *Processor) NextPriceUp(prices ...items_types.PriceType) items_types.PriceType {
	price := items_types.PriceType(0.0)
	if len(prices) == 0 {
		price = pp.GetCurrentPrice()
	} else {
		price = prices[0]
	}
	return price * items_types.PriceType(1+pp.GetDeltaPrice()/100)
}

func (pp *Processor) NextPriceDown(prices ...items_types.PriceType) items_types.PriceType {
	price := items_types.PriceType(0.0)
	if len(prices) == 0 {
		price = pp.GetCurrentPrice()
	} else {
		price = prices[0]
	}
	return price * items_types.PriceType(1-pp.GetDeltaPrice()/100)
}

func (pp *Processor) NextQuantityUp(quantity items_types.QuantityType) items_types.QuantityType {
	return quantity * items_types.QuantityType(1+pp.GetDeltaQuantity()/100)
}

func (pp *Processor) NextQuantityDown(quantity items_types.QuantityType) items_types.QuantityType {
	return quantity * items_types.QuantityType(1-pp.GetDeltaQuantity())
}
