package streams

import (
	"github.com/adshao/go-binance/v2"
)

func GetFilledOrderHandler() (executeOrderChan chan *binance.WsUserDataEvent) {
	executeOrderChan = make(chan *binance.WsUserDataEvent, 1)
	go func() {
		userDataChannel, err := GetUserDataChannel()
		if !err {
			return
		}
		for {
			event := <-userDataChannel
			if event.Event == binance.UserDataEventTypeExecutionReport &&
				(event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
					event.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled)) {
				executeOrderChan <- event
			}
		}
	}()
	return
}
