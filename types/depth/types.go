package depth

import (
	"sync"

	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

type (
	Depths struct {
		symbol string
		degree int
		asks   *depths_types.Asks
		// asks              *btree.BTree
		// asksCountQuantity int
		// asksSummaQuantity types.QuantityType
		// asksMinMax        *btree.BTree
		// askNormalized     *btree.BTree
		bids *depths_types.Bids
		// bids              *btree.BTree
		// bidsCountQuantity int
		// bidsSummaQuantity types.QuantityType
		// bidsMinMax        *btree.BTree
		// bidNormalized     *btree.BTree
		mutex           *sync.Mutex
		LastUpdateID    int64
		limitDepth      depths_types.DepthAPILimit
		limitStream     depths_types.DepthStreamLevel
		rateStream      depths_types.DepthStreamRate
		percentToTarget float64
		expBase         int
	}
)

// Lock implements depth_interface.Depths.
func (d *Depths) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Depths) Unlock() {
	d.mutex.Unlock()
}

// TryLock implements depth_interface.Depths.
func (d *Depths) TryLock() bool {
	return d.mutex.TryLock()
}

func (d *Depths) GetAsks() *depths_types.Asks {
	return d.asks
}

func (d *Depths) GetBids() *depths_types.Bids {
	return d.bids
}

func (a *Depths) UpdateAsk(item *items_types.Ask) bool {
	a.bids.Delete(items_types.NewBid(item.GetDepthItem().GetPrice(), item.GetDepthItem().GetQuantity()))
	return a.asks.Update(item)
}

func (a *Depths) UpdateBid(item *items_types.Bid) bool {
	a.asks.Delete(items_types.NewAsk(item.GetDepthItem().GetPrice(), item.GetDepthItem().GetQuantity()))
	return a.bids.Update(item)
}

func (d *Depths) GetPercentToTarget() float64 {
	return d.percentToTarget
}
