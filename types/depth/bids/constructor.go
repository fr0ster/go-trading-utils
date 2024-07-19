package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
)

func New(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth depths_types.DepthAPILimit,
	expBase int,
	rate ...depths_types.DepthStreamRate) *Bids {
	return &Bids{tree: depths_types.New(degree, symbol, targetPercent, limitDepth, expBase, rate...)}
}
