package depth

import (
	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

// Symbol implements depth_interface.Depths.
func (d *Depths) Symbol() string {
	return d.symbol
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

func (d *Depths) NextPriceUp(percent items_types.PricePercentType, price ...items_types.PriceType) (next items_types.PriceType) {
	var asksFilter items_types.DepthFilter
	if len(price) > 0 {
		asksFilter = func(i *items_types.DepthItem) bool {
			return i.GetPrice() > price[0]
		}
		if val := d.asks.GetFiltered(asksFilter); val != nil {
			next = val.NextPriceUp(percent * d.GetNextUpCoefficient())
		}
	} else {
		next = d.asks.NextPriceUp(percent * d.GetNextUpCoefficient())
	}
	return
}

func (d *Depths) NextPriceDown(percent items_types.PricePercentType, price ...items_types.PriceType) (next items_types.PriceType) {
	var bidsFilter items_types.DepthFilter
	if len(price) > 0 {
		bidsFilter = func(i *items_types.DepthItem) bool {
			return i.GetPrice() < price[0]
		}
		if val := d.bids.GetFiltered(bidsFilter); val != nil {
			next = val.NextPriceDown(percent * d.GetNextDownCoefficient())

		}
	} else {
		next = d.bids.NextPriceDown(percent * d.GetNextDownCoefficient())
	}
	return
}

func (d *Depths) SetStartDepthStream(startDepthStreamCreator func(*Depths) types.StreamFunction) {
	if startDepthStreamCreator != nil {
		d.startDepthStream = startDepthStreamCreator(d)
	}
}

func (d *Depths) SetInit(initCreator func(*Depths) types.InitFunction) {
	if initCreator != nil {
		d.Init = initCreator(d)
	}
}
