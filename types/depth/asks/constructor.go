package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
)

func New(degree int, symbol string) *Asks {
	return &Asks{tree: depths_types.New(degree, symbol)}
}

// Symbol implements depth_interface.Depths.
func (d *Asks) Symbol() string {
	return d.tree.Symbol()
}

func (d *Asks) Degree() int {
	return d.tree.Degree()
}
