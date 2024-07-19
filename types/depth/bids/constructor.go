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

// Symbol implements depth_interface.Depths.
func (d *Bids) Symbol() string {
	return d.tree.Symbol()
}

func (d *Bids) Degree() int {
	return d.tree.Degree()
}

func (d *Bids) ExpBase() int {
	return d.tree.ExpBase()
}

func (d *Bids) TargetPercent() float64 {
	return d.tree.TargetPercent()
}

func (d *Bids) LimitDepth() depths_types.DepthAPILimit {
	return d.tree.LimitDepth()
}

func (d *Bids) LimitStream() depths_types.DepthStreamLevel {
	return d.tree.LimitStream()
}

func (d *Bids) RateStream() depths_types.DepthStreamRate {
	return d.tree.RateStream()
}
