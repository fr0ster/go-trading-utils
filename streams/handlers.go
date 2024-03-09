package streams

import "github.com/adshao/go-binance/v2"

func GetFilledOrderHandler() (func(event *binance.WsUserDataEvent), chan *binance.WsUserDataEvent) {
	executeOrderChan := make(chan *binance.WsUserDataEvent, 1)
	return func(event *binance.WsUserDataEvent) {
		if event.Event == binance.UserDataEventTypeExecutionReport && event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) {
			executeOrderChan <- event
		}
	}, executeOrderChan
}
