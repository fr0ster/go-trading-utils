package depth

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Symbol implements depth_interface.Depths.
func (d *Depths) Symbol() string {
	return d.symbol
}

func (d *Depths) GetLimitDepth() depths_types.DepthAPILimit {
	return d.limitDepth
}

func (d *Depths) GetLimitStream() depths_types.DepthStreamLevel {
	return d.limitStream
}

func (d *Depths) GetRateStream() depths_types.DepthStreamRate {
	return d.rateStream
}

func (d *Depths) GetNextUpCoefficient() items_types.PricePercentType {
	coefficients := items_types.PriceType(d.GetAsks().GetSummaValue() / d.GetBids().GetSummaValue())
	if coefficients > 1 {
		return items_types.PricePercentType(coefficients)
	} else {
		return 1
	}
}

func (d *Depths) GetNextDownCoefficient() items_types.PricePercentType {
	coefficients := items_types.PriceType(d.GetAsks().GetSummaValue() / d.GetBids().GetSummaValue())
	if coefficients > 1 {
		return items_types.PricePercentType(coefficients)
	} else {
		return 1
	}
}

func (d *Depths) NextPriceUp(percent items_types.PricePercentType, price ...items_types.PriceType) items_types.PriceType {
	var asksFilter items_types.DepthFilter
	if len(price) > 0 {
		asksFilter = func(i *items_types.DepthItem) bool {
			return i.GetPrice() > price[0]
		}
		if val := d.asks.GetFiltered(asksFilter); val != nil {
			return val.NextPriceUp(percent * d.GetNextUpCoefficient())
		}
	}
	return d.asks.NextPriceUp(percent)
}

func (d *Depths) NextPriceDown(percent items_types.PricePercentType, price ...items_types.PriceType) items_types.PriceType {
	var bidsFilter items_types.DepthFilter
	if len(price) > 0 {
		bidsFilter = func(i *items_types.DepthItem) bool {
			return i.GetPrice() < price[0]
		}
		if val := d.bids.GetFiltered(bidsFilter); val != nil {
			return val.NextPriceDown(percent * d.GetNextDownCoefficient())
		}
	}
	return d.bids.NextPriceDown(percent)
}
