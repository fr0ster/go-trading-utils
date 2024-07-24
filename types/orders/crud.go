package orders

import "github.com/fr0ster/go-trading-utils/types"

func (o *Orders) SetStartUserDataStream(startUserDataStreamCreator func(*Orders) types.StreamFunction) {
	if startUserDataStreamCreator != nil {
		o.startUserDataStream = startUserDataStreamCreator(o)
	}
}

func (o *Orders) SetOrderCreator(createOrderCreator func(*Orders) CreateOrderFunction) {
	if createOrderCreator != nil {
		o.CreateOrder = createOrderCreator(o)
	}
}

func (o *Orders) SetGetOpenOrders(getOpenOrders func(*Orders) func() ([]*Order, error)) {
	if getOpenOrders != nil {
		o.GetOpenOrders = getOpenOrders(o)
	}
}

func (o *Orders) SetGetAllOrders(getAllOrders func(*Orders) func() ([]*Order, error)) {
	if getAllOrders != nil {
		o.GetAllOrders = getAllOrders(o)
	}
}

func (o *Orders) SetGetOrder(getOrder func(*Orders) func(orderID int64) (*Order, error)) {
	if getOrder != nil {
		o.GetOrder = getOrder(o)
	}
}

func (o *Orders) SetCancelOrder(cancelOrder func(*Orders) func(orderID int64) (*CancelOrderResponse, error)) {
	if cancelOrder != nil {
		o.CancelOrder = cancelOrder(o)
	}
}

func (o *Orders) SetCancelAllOrders(cancelAllOrders func(*Orders) func() (err error)) {
	if cancelAllOrders != nil {
		o.CancelAllOrders = cancelAllOrders(o)
	}
}
