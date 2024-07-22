package bids

import depths_types "github.com/fr0ster/go-trading-utils/types/depths/depths"

type (
	Bids struct{ tree *depths_types.Depths }
)
