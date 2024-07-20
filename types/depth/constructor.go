package depth

import (
	"errors"
	"sync"

	asks_types "github.com/fr0ster/go-trading-utils/types/depth/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depth/bids"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(
	degree int,
	symbol string,
	isMinMax bool,
	targetPercent float64,
	limitDepth depths_types.DepthAPILimit,
	expBase int,
	rate ...depths_types.DepthStreamRate) *Depths {
	var (
		limitStream depths_types.DepthStreamLevel
		rateStream  depths_types.DepthStreamRate
		// asksMinMax  *btree.BTree
		// bidsMinMax  *btree.BTree
	)
	switch limitDepth {
	case depths_types.DepthAPILimit5:
		limitStream = depths_types.DepthStreamLevel5
	case depths_types.DepthAPILimit10:
		limitStream = depths_types.DepthStreamLevel10
	default:
		limitStream = depths_types.DepthStreamLevel20
	}
	if len(rate) == 0 {
		rateStream = depths_types.DepthStreamRate100ms
	} else {
		rateStream = rate[0]
	}
	// if isMinMax {
	// 	asksMinMax = btree.New(degree)
	// 	bidsMinMax = btree.New(degree)
	// }
	return &Depths{
		symbol: symbol,
		degree: degree,
		asks:   asks_types.New(degree, symbol, targetPercent, limitDepth, expBase, rate...),
		// asksMinMax:      asksMinMax,
		// askNormalized:   btree.New(degree),
		bids: bids_types.New(degree, symbol, targetPercent, limitDepth, expBase, rate...),
		// bidsMinMax:      bidsMinMax,
		// bidNormalized:   btree.New(degree),
		mutex:           &sync.Mutex{},
		limitDepth:      limitDepth,
		limitStream:     limitStream,
		rateStream:      rateStream,
		percentToTarget: targetPercent,
		expBase:         expBase,
	}
}

func Binance2BookTicker(binanceDepth interface{}) (*items_types.DepthItem, error) {
	switch binanceDepth := binanceDepth.(type) {
	case *items_types.DepthItem:
		return binanceDepth, nil
	}
	return nil, errors.New("it's not a types.DepthItemType")
}

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

func (d *Depths) GetNextUpCoefficient() items_types.PriceType {
	coefficients := items_types.PriceType(d.GetAsks().GetSummaValue() / d.GetBids().GetSummaValue())
	if coefficients > 1 {
		return coefficients
	} else {
		return 1
	}
}

func (d *Depths) GetNextDownCoefficient() items_types.PriceType {
	coefficients := items_types.PriceType(d.GetAsks().GetSummaValue() / d.GetBids().GetSummaValue())
	if coefficients > 1 {
		return coefficients
	} else {
		return 1
	}
}

func (d *Depths) NextPriceUp(percent float64, price ...items_types.PriceType) items_types.PriceType {
	var asksFilter items_types.DepthFilter
	if len(price) > 0 {
		asksFilter = func(i *items_types.DepthItem) bool {
			return i.GetPrice() > price[0]
		}
		if val := d.asks.GetFiltered(asksFilter); val != nil {
			return val.NextPriceUp(percent)
		} else {
			return price[0] * items_types.PriceType(1+percent)
		}
	}
	return d.asks.NextPriceUp(percent)
}

func (d *Depths) NextPriceDown(percent float64, price ...items_types.PriceType) items_types.PriceType {
	var bidsFilter items_types.DepthFilter
	if len(price) > 0 {
		bidsFilter = func(i *items_types.DepthItem) bool {
			return i.GetPrice() < price[0]
		}
		if val := d.bids.GetFiltered(bidsFilter); val != nil {
			return val.NextPriceDown(percent)
		} else {
			return price[0] * items_types.PriceType(1-percent)
		}
	}
	return d.bids.NextPriceDown(percent)
}
