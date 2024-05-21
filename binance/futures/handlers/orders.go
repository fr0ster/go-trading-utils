package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
)

func GetChangingOfOrdersGuard(
	source chan *futures.WsUserDataEvent,
	statuses []futures.OrderStatusType) (out chan *futures.WsUserDataEvent) {
	out = make(chan *futures.WsUserDataEvent)
	go func() {
		for {
			event := <-source
			if event.Event == futures.UserDataEventTypeOrderTradeUpdate {
				for _, status := range statuses {
					if event.OrderTradeUpdate.Status == status {
						out <- event
					}
				}
			}
		}
	}()
	return
}
