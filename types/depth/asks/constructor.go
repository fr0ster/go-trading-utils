package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
)

func New(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth depths_types.DepthAPILimit,
	expBase int,
	rate ...depths_types.DepthStreamRate) *Asks {
	return &Asks{tree: depths_types.New(degree, symbol, targetPercent, limitDepth, expBase, rate...)}
}

// Symbol implements depth_interface.Depths.
func (d *Asks) Symbol() string {
	return d.tree.Symbol()
}

func (d *Asks) Degree() int {
	return d.tree.Degree()
}

func (d *Asks) ExpBase() int {
	return d.tree.ExpBase()
}

func (d *Asks) TargetPercent() float64 {
	return d.tree.TargetPercent()
}

func (d *Asks) LimitDepth() depths_types.DepthAPILimit {
	return d.tree.LimitDepth()
}

func (d *Asks) LimitStream() depths_types.DepthStreamLevel {
	return d.tree.LimitStream()
}

func (d *Asks) RateStream() depths_types.DepthStreamRate {
	return d.tree.RateStream()
}
