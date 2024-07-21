package depth

import (
	"sync"
	"time"

	asks_types "github.com/fr0ster/go-trading-utils/types/depth/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depth/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

type (
	Depths struct {
		symbol string
		degree int
		asks   *asks_types.Asks
		// asks              *btree.BTree
		// asksCountQuantity int
		// asksSummaQuantity types.QuantityType
		// asksMinMax        *btree.BTree
		// askNormalized     *btree.BTree
		bids *bids_types.Bids
		// bids              *btree.BTree
		// bidsCountQuantity int
		// bidsSummaQuantity types.QuantityType
		// bidsMinMax        *btree.BTree
		// bidNormalized     *btree.BTree
		mutex        *sync.Mutex
		LastUpdateID int64

		stop             chan struct{}
		resetEvent       chan error
		timeOut          time.Duration
		StartDepthStream func() (chan struct{}, chan struct{}, error)
		Init             func(*Depths) error
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

func (d *Depths) GetAsks() *asks_types.Asks {
	return d.asks
}

func (d *Depths) GetBids() *bids_types.Bids {
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

func (a *Depths) ResetEvent(err error) {
	a.resetEvent <- err
}
