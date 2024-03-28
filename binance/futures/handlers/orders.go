package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
)

func GetChangingOfOrdersGuard(
	source chan *futures.WsUserDataEvent,
	eventCheck futures.UserDataEventType,
	statuses []futures.OrderStatusType) (out chan *futures.WsUserDataEvent) {
	out = make(chan *futures.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == eventCheck {
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
