package processor

import (
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *Processor) GetDepth() *depth_types.Depths {
	return pp.depth
}

func (pp *Processor) SetDepth(depth *depth_types.Depths) {
	pp.depth = depth
}

func (pp *Processor) NextPriceUp(prices ...items_types.PriceType) items_types.PriceType {
	var err error
	if pp.depth != nil {
		return pp.RoundPrice(pp.GetDepth().NextPriceUp(items_types.PricePercentType(pp.GetDeltaPrice()), prices...))
	} else {
		price := items_types.PriceType(0.0)
		if len(prices) == 0 {
			price, err = pp.GetCurrentPrice()
			if err != nil {
				return 0
			}
		} else {
			price = prices[0]
		}
		return pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
	}
}

func (pp *Processor) NextPriceDown(prices ...items_types.PriceType) items_types.PriceType {
	var err error
	if pp.depth != nil {
		return pp.RoundPrice(pp.GetDepth().NextPriceDown(items_types.PricePercentType(pp.GetDeltaPrice()), prices...))
	} else {
		price := items_types.PriceType(0.0)
		if len(prices) == 0 {
			price, err = pp.GetCurrentPrice()
			if err != nil {
				return 0
			}
		} else {
			price = prices[0]
		}
		return pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
	}
}

func (pp *Processor) NextQuantityUp(quantity items_types.QuantityType) items_types.QuantityType {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *Processor) NextQuantityDown(quantity items_types.QuantityType) items_types.QuantityType {
	return pp.RoundQuantity(quantity * (1 - pp.GetDeltaQuantity()))
}
