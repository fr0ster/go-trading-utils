package bids

import depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"

type (
	Bids struct{ tree *depths_types.Depths }
)
