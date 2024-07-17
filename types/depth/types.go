package depth

import (
	"sync"
	"time"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
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
	Depth struct {
		symbol            string
		degree            int
		asks              *btree.BTree
		asksCountQuantity int
		asksSummaQuantity types.QuantityType
		asksMinMax        *btree.BTree
		// askNormalized     *btree.BTree
		bids              *btree.BTree
		bidsCountQuantity int
		bidsSummaQuantity types.QuantityType
		bidsMinMax        *btree.BTree
		// bidNormalized     *btree.BTree
		mutex           *sync.Mutex
		LastUpdateID    int64
		limitDepth      DepthAPILimit
		limitStream     DepthStreamLevel
		rateStream      DepthStreamRate
		percentToTarget float64
		expBase         int
	}
)

// Lock implements depth_interface.Depths.
func (d *Depth) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Depth) Unlock() {
	d.mutex.Unlock()
}

// TryLock implements depth_interface.Depths.
func (d *Depth) TryLock() bool {
	return d.mutex.TryLock()
}
