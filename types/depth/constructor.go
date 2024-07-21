package depth

import (
	"errors"
	"sync"
	"time"

	asks_types "github.com/fr0ster/go-trading-utils/types/depth/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depth/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(
	degree int,
	symbol string,
	timeOut time.Duration,
	startDepthStreamCreator func(*Depths) func() (chan struct{}, chan struct{}, error),
	initCreator func(*Depths) func() (err error),
	stops ...chan struct{}) *Depths {
	var (
		stop chan struct{}
	)
	if len(stops) > 0 {
		stop = stops[0]
	} else {
		stop = make(chan struct{}, 1)
	}
	this := &Depths{
		symbol:           symbol,
		degree:           degree,
		asks:             asks_types.New(degree, symbol),
		bids:             bids_types.New(degree, symbol),
		mutex:            &sync.Mutex{},
		stop:             stop,
		resetEvent:       make(chan error, 1),
		timeOut:          timeOut,
		StartDepthStream: nil,
		Init:             nil,
	}
	if startDepthStreamCreator != nil && initCreator != nil {
		this.StartDepthStream = startDepthStreamCreator(this)
		this.Init = initCreator(this)
	}
	return this
}

func Binance2BookTicker(binanceDepth interface{}) (*items_types.DepthItem, error) {
	switch binanceDepth := binanceDepth.(type) {
	case *items_types.DepthItem:
		return binanceDepth, nil
	}
	return nil, errors.New("it's not a types.DepthItemType")
}
