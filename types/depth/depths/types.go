package depths

import (
	"sync"
	"time"

	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

const (
	DepthStreamLevel5    DepthStreamLevel = 5
	DepthStreamLevel10   DepthStreamLevel = 10
	DepthStreamLevel20   DepthStreamLevel = 20
	DepthAPILimit5       DepthAPILimit    = 5
	DepthAPILimit10      DepthAPILimit    = 10
	DepthAPILimit20      DepthAPILimit    = 20
	DepthAPILimit50      DepthAPILimit    = 50
	DepthAPILimit100     DepthAPILimit    = 100
	DepthAPILimit500     DepthAPILimit    = 500
	DepthAPILimit1000    DepthAPILimit    = 1000
	DepthStreamRate100ms DepthStreamRate  = DepthStreamRate(100 * time.Millisecond)
	DepthStreamRate250ms DepthStreamRate  = DepthStreamRate(250 * time.Millisecond)
	DepthStreamRate500ms DepthStreamRate  = DepthStreamRate(500 * time.Millisecond)
)

type (
	DepthStreamLevel int
	DepthAPILimit    int
	DepthStreamRate  time.Duration
)

type (
	Asks   struct{ tree *Depths }
	Bids   struct{ tree *Depths }
	Depths struct {
		symbol        string
		degree        int
		tree          *btree.BTree
		mutex         *sync.Mutex
		countQuantity int
		summaQuantity types.QuantityType
		limitStream   DepthStreamLevel
		rateStream    DepthStreamRate
	}
)
