package analyzer

import (
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	DepthAnalyzer interface {
		Update(depth_interface.Depth) error
		GetLevels(side types.DepthSide) *btree.BTree
		Lock()
		Unlock()
	}
)
