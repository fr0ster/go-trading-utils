package handlers

import (
	"github.com/adshao/go-binance/v2"
)

func GetChangingOfOrdersGuard(
	source chan *binance.WsUserDataEvent,
	statuses []binance.OrderStatusType) (out chan *binance.WsUserDataEvent) {
	out = make(chan *binance.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == binance.UserDataEventTypeExecutionReport {
				for _, status := range statuses {
					if event.OrderUpdate.Status == string(status) {
						out <- event
					}
				}
			}
			source <- event
		}
	}()
	return
}
