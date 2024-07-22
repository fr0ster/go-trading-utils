package depth

import (
	"sync"
	"time"

	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
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
	Depths struct {
		symbol       string
		degree       int
		asks         *asks_types.Asks
		bids         *bids_types.Bids
		mutex        *sync.Mutex
		LastUpdateID int64

		stop             chan struct{}
		resetEvent       chan error
		timeOut          time.Duration
		startDepthStream func() (chan struct{}, chan struct{}, error)
		Init             func() (err error)
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
