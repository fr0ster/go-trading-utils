package depth

import (
	"errors"
	"sync"

	"github.com/google/btree"
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree int, symbol string, limitDepth DepthAPILimit, rate ...DepthStreamRate) *Depth {
	var (
		limitStream DepthStreamLevel
		rateStream  DepthStreamRate
	)
	switch limitDepth {
	case DepthAPILimit5:
		limitStream = DepthStreamLevel5
	case DepthAPILimit10:
		limitStream = DepthStreamLevel10
	default:
		limitStream = DepthStreamLevel20
	}
	if len(rate) == 0 {
		rateStream = DepthStreamRate100ms
	} else {
		rateStream = rate[0]
	}
	return &Depth{
		symbol:      symbol,
		degree:      degree,
		asks:        btree.New(degree),
		asksMinMax:  btree.New(degree),
		bids:        btree.New(degree),
		bidsMinMax:  btree.New(degree),
		mutex:       &sync.Mutex{},
		limitDepth:  limitDepth,
		limitStream: limitStream,
		rateStream:  rateStream,
	}
}

func Binance2BookTicker(binanceDepth interface{}) (*DepthItem, error) {
	switch binanceDepth := binanceDepth.(type) {
	case *DepthItem:
		return binanceDepth, nil
	}
	return nil, errors.New("it's not a DepthItemType")
}

// Symbol implements depth_interface.Depths.
func (d *Depth) Symbol() string {
	return d.symbol
}

func (d *Depth) GetLimitDepth() DepthAPILimit {
	return d.limitDepth
}

func (d *Depth) GetLimitStream() DepthStreamLevel {
	return d.limitStream
}

func (d *Depth) GetRateStream() DepthStreamRate {
	return d.rateStream
}

func (d *Depth) GetAsksSummaQuantity() float64 {
	return d.asksSummaQuantity
}

func (d *Depth) GetBidsSummaQuantity() float64 {
	return d.bidsSummaQuantity
}
