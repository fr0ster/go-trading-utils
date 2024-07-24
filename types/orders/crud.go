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
