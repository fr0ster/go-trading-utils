package depths

import (
	"sync"

	"github.com/google/btree"
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth DepthAPILimit,
	expBase int,
	rate ...DepthStreamRate) *Depths {
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
	return &Depths{
		symbol:        symbol,
		degree:        degree,
		expBase:       expBase,
		targetPercent: targetPercent,
		limitDepth:    limitDepth,
		limitStream:   limitStream,
		rateStream:    rateStream,

		tree:  btree.New(degree),
		mutex: &sync.Mutex{},

		countQuantity: 0,
		summaQuantity: 0,
		summaValue:    0,
	}
}

// Depths -
func (d *Depths) Symbol() string {
	return d.symbol
}

func (d *Depths) Degree() int {
	return d.degree
}

func (d *Depths) ExpBase() int {
	return d.expBase
}

func (d *Depths) TargetPercent() float64 {
	return d.targetPercent
}

func (d *Depths) LimitDepth() DepthAPILimit {
	return d.limitDepth
}

func (d *Depths) LimitStream() DepthStreamLevel {
	return d.limitStream
}

func (d *Depths) RateStream() DepthStreamRate {
	return d.rateStream
}
