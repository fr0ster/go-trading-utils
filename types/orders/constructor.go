package orders

import (
	"time"

	"github.com/fr0ster/go-trading-utils/types"
)

func (o *Orders) ResetEvent(err error) {
	o.resetEvent <- err
}

func (o *Orders) Symbol() string {
	return o.symbol
}

func New(
	symbol string,
	startUserDataStreamCreator func(*Orders) types.StreamFunction,
	createOrderCreator func(*Orders) CreateOrderFunction,
	openOrdersCreator func(*Orders) func() ([]*Order, error),
	allOrdersCreator func(*Orders) func() ([]*Order, error),
	getOrderCreator func(*Orders) func(orderID int64) (*Order, error),
	cancelOrderCreator func(*Orders) func(orderID int64) (*CancelOrderResponse, error),
	cancelAllOrdersCreator func(*Orders) func() (err error),
	stops ...chan struct{}) (this *Orders) {
	var stop chan struct{}
	if len(stops) > 0 {
		stop = stops[0]
	} else {
		stop = make(chan struct{})
	}
	this = &Orders{
		symbol:     symbol,
		stop:       stop,
		resetEvent: make(chan error),
		timeOut:    1 * time.Hour,
	}
	this.SetStartUserDataStream(startUserDataStreamCreator)
	this.SetOrderCreator(createOrderCreator)
	this.SetGetOpenOrders(openOrdersCreator)
	this.SetGetAllOrders(allOrdersCreator)
	this.SetGetOrder(getOrderCreator)
	this.SetCancelOrder(cancelOrderCreator)
	return
}
