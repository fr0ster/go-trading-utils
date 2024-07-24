package depth

import (
	"sync"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(
	degree int,
	symbol string,
	startDepthStreamCreator func(*Depths) types.StreamFunction,
	initCreator func(*Depths) types.InitFunction,
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
		timeOut:          1 * time.Hour,
		startDepthStream: nil,
		Init:             nil,
	}
	this.SetStartDepthStream(startDepthStreamCreator)
	this.SetInit(initCreator)
	return this
}
