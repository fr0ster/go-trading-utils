package analyzer

import (
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/google/btree"
)

type (
	DepthAnalyzer interface {
		Update(depth_interface.Depth)
		GetLevels() *btree.BTree
	}
)
