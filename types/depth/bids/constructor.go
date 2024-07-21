package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
)

func New(degree int, symbol string) *Bids {
	return &Bids{tree: depths_types.New(degree, symbol)}
}

// Symbol implements depth_interface.Depths.
func (d *Bids) Symbol() string {
	return d.tree.Symbol()
}

func (d *Bids) Degree() int {
	return d.tree.Degree()
}
