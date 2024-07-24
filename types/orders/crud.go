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

func (o *Orders) SetGetOpenOrders(getOpenOrders func(*Orders) OpenOrderFunction) {
	if getOpenOrders != nil {
		o.GetOpenOrders = getOpenOrders(o)
	}
}

func (o *Orders) SetGetAllOrders(getAllOrders func(*Orders) AllOrdersFunction) {
	if getAllOrders != nil {
		o.GetAllOrders = getAllOrders(o)
	}
}

func (o *Orders) SetGetOrder(getOrder func(*Orders) GetOrderFunction) {
	if getOrder != nil {
		o.GetOrder = getOrder(o)
	}
}

func (o *Orders) SetCancelOrder(cancelOrder func(*Orders) CancelOrderFunction) {
	if cancelOrder != nil {
		o.CancelOrder = cancelOrder(o)
	}
}

func (o *Orders) SetCancelAllOrders(cancelAllOrders func(*Orders) CancelAllOrdersFunction) {
	if cancelAllOrders != nil {
		o.CancelAllOrders = cancelAllOrders(o)
	}
}
