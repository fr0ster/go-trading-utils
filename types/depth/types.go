package depth

import (
	"sync"
	"time"

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
	DeviationItem struct {
		Price     float64
		Quantity  float64
		Deviation float64
	}
	QuantityItem struct {
		Quantity float64
		Depths   *btree.BTree
	}
	DepthItem struct {
		Price    float64
		Quantity float64
	}
	// DepthItemType - тип для зберігання заявок в стакані
	Depth struct {
		symbol            string
		degree            int
		asks              *btree.BTree
		asksCountQuantity int
		asksSummaQuantity float64
		asksMinMax        *btree.BTree
		bids              *btree.BTree
		bidsCountQuantity int
		bidsSummaQuantity float64
		bidsMinMax        *btree.BTree
		mutex             *sync.Mutex
		LastUpdateID      int64
		limitDepth        DepthAPILimit
		limitStream       DepthStreamLevel
		rateStream        DepthStreamRate
		percentToTarget   float64
		expBase           int
	}
	DepthFilter func(*DepthItem) bool
	DepthTester func(result *DepthItem, target *DepthItem) bool
)

func (i *DepthItem) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItem).Price
}

func (i *DepthItem) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItem).Price
}

// GetAskDeviation implements depth_interface.Depths.
func (d *DepthItem) GetQuantityDeviation(middle float64) float64 {
	return d.Quantity - middle
}

func (i *QuantityItem) Less(than btree.Item) bool {
	return i.Quantity < than.(*QuantityItem).Quantity
}

func (i *QuantityItem) Equal(than btree.Item) bool {
	return i.Quantity == than.(*QuantityItem).Quantity
}

// Lock implements depth_interface.Depths.
func (d *Depth) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Depth) Unlock() {
	d.mutex.Unlock()
}
