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
		tree:          btree.New(degree),
		mutex:         &sync.Mutex{},
		countQuantity: 0,
		summaQuantity: 0,
		limitStream:   limitStream,
		rateStream:    rateStream,
	}
}
