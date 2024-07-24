package orders

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	"github.com/fr0ster/go-trading-utils/types"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
)

func UserDataStreamCreator(
	client *futures.Client,
	handlerCreator func(d *orders_types.Orders) futures.WsUserDataHandler,
	errHandlerCreator func(d *orders_types.Orders) futures.ErrHandler) func(d *orders_types.Orders) types.StreamFunction {
	return func(o *orders_types.Orders) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Отримуємо новий або той же самий ключ для прослуховування подій користувача при втраті з'єднання
			listenKey, err := client.NewStartUserStreamService().Do(context.Background())
			if err != nil {
				return
			}
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsUserDataServe(listenKey, handlerCreator(o), errHandlerCreator(o))
			return
		}
	}
}

func CallBackCreator(
	handlers ...func(d *orders_types.Orders) futures.WsUserDataHandler) func(d *orders_types.Orders) futures.WsUserDataHandler {
	return func(d *orders_types.Orders) futures.WsUserDataHandler {
		var stack []futures.WsUserDataHandler
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *futures.WsUserDataEvent) {
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func WsErrorHandlerCreator(handlers ...func(*orders_types.Orders) futures.ErrHandler) func(*orders_types.Orders) futures.ErrHandler {
	return func(o *orders_types.Orders) futures.ErrHandler {
		var stack []futures.ErrHandler
		for _, handler := range handlers {
			stack = append(stack, handler(o))
		}
		return func(err error) {
			logrus.Errorf("Spot wsErrorHandler error: %v", err)
			o.ResetEvent(err)
			for _, handler := range stack {
				handler(err)
			}
		}
	}
}
