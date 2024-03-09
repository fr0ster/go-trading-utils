package streams

import "github.com/adshao/go-binance/v2"

func GetFilledOrderHandler(executeOrderChan chan *binance.WsUserDataEvent) func(event *binance.WsUserDataEvent) {
	return func(event *binance.WsUserDataEvent) {
		if event.Event == binance.UserDataEventTypeExecutionReport && event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) {
			executeOrderChan <- event
		}
	}
}
