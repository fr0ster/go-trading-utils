package orders

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	"github.com/fr0ster/go-trading-utils/types"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
)

func UserDataStreamCreator(
	client *binance.Client,
	handlerCreator func(d *orders_types.Orders) binance.WsUserDataHandler,
	errHandlerCreator func(d *orders_types.Orders) binance.ErrHandler) func(d *orders_types.Orders) types.StreamFunction {
	return func(o *orders_types.Orders) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Отримуємо новий або той же самий ключ для прослуховування подій користувача при втраті з'єднання
			listenKey, err := client.NewStartUserStreamService().Do(context.Background())
			if err != nil {
				return
			}
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsUserDataServe(listenKey, handlerCreator(o), errHandlerCreator(o))
			return
		}
	}
}

func CallBackCreator(
	handlers ...func(d *orders_types.Orders) binance.WsUserDataHandler) func(d *orders_types.Orders) binance.WsUserDataHandler {
	return func(d *orders_types.Orders) binance.WsUserDataHandler {
		var stack []binance.WsUserDataHandler
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *binance.WsUserDataEvent) {
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func WsErrorHandlerCreator(handlers ...func(*orders_types.Orders) binance.ErrHandler) func(*orders_types.Orders) binance.ErrHandler {
	return func(o *orders_types.Orders) binance.ErrHandler {
		var stack []binance.ErrHandler
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
