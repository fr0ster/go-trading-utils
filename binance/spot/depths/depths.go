package depths

import (
	"context"

	"github.com/adshao/go-binance/v2"

	"github.com/fr0ster/go-trading-utils/types"
	depths_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"

	"github.com/sirupsen/logrus"
)

func InitCreator(limit depths_types.DepthAPILimit, client *binance.Client) func(d *depths_types.Depths) types.InitFunction {
	return func(d *depths_types.Depths) types.InitFunction {
		return func() (err error) {
			res, err :=
				client.NewDepthService().
					Symbol(string(d.Symbol())).
					Limit(int(limit)).
					Do(context.Background())
			if err != nil {
				return err
			}
			d.GetBids().Clear()
			for _, bid := range res.Bids {
				price, quantity, _ := bid.Parse()
				d.GetBids().Update(items_types.NewBid(items_types.PriceType(price), items_types.QuantityType(quantity)))
			}
			d.GetAsks().Clear()
			for _, ask := range res.Asks {
				price, quantity, _ := ask.Parse()
				d.GetAsks().Update(items_types.NewAsk(items_types.PriceType(price), items_types.QuantityType(quantity)))
			}
			d.LastUpdateID = res.LastUpdateID
			return nil
		}
	}
}

func DepthStreamCreator(
	handlerCreator func(d *depths_types.Depths) binance.WsDepthHandler,
	errHandlerCreator func(d *depths_types.Depths) binance.ErrHandler) func(d *depths_types.Depths) types.StreamFunction {
	return func(d *depths_types.Depths) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsDepthServe100Ms(d.Symbol(), handlerCreator(d), errHandlerCreator(d))
			return
		}
	}
}

func eventHandlerCreator(d *depths_types.Depths) binance.WsDepthHandler {
	return func(event *binance.WsDepthEvent) {
		func() {
			d.Lock()         // Locking the depths
			defer d.Unlock() // Unlocking the depths
			if event.LastUpdateID < d.LastUpdateID {
				return
			}
			if event.LastUpdateID != int64(d.LastUpdateID)+1 {
				d.Init()
			} else if event.LastUpdateID == int64(d.LastUpdateID)+1 {
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						return
					}
					d.GetBids().Update(items_types.NewBid(items_types.PriceType(price), items_types.QuantityType(quantity)))
				}
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						return
					}
					d.GetAsks().Update(items_types.NewAsk(items_types.PriceType(price), items_types.QuantityType(quantity)))
				}
				d.LastUpdateID = event.LastUpdateID
			}
		}()
	}
}

func CallBackCreator(
	handlers ...func(d *depths_types.Depths) binance.WsDepthHandler) func(d *depths_types.Depths) binance.WsDepthHandler {
	return func(d *depths_types.Depths) binance.WsDepthHandler {
		d.Init()
		var stack []binance.WsDepthHandler
		d.Init()
		standardHandlers := eventHandlerCreator(d)
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *binance.WsDepthEvent) {
			standardHandlers(event)
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func PartialDepthStreamCreator(
	levels depths_types.DepthStreamLevel,
	handlerCreator func(d *depths_types.Depths) binance.WsPartialDepthHandler,
	errHandlerCreator func(d *depths_types.Depths) binance.ErrHandler) func(d *depths_types.Depths) types.StreamFunction {
	return func(d *depths_types.Depths) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsPartialDepthServe100Ms(d.Symbol(), string(rune(levels)), handlerCreator(d), errHandlerCreator(d))
			return
		}
	}
}

func PartialDepthHandlerCreator(d *depths_types.Depths) binance.WsPartialDepthHandler {
	return func(event *binance.WsPartialDepthEvent) {
		func() {
			d.Lock()         // Locking the depths
			defer d.Unlock() // Unlocking the depths
			if event.LastUpdateID < d.LastUpdateID {
				return
			}
			if event.LastUpdateID != int64(d.LastUpdateID)+1 {
				d.Init()
			} else if event.LastUpdateID == int64(d.LastUpdateID)+1 {
				logrus.Debugf("Spot %v depth update event received", d.Symbol())
				for _, bid := range event.Bids {
					price, quantity, err := bid.Parse()
					if err != nil {
						return
					}
					d.GetBids().Update(items_types.NewBid(items_types.PriceType(price), items_types.QuantityType(quantity)))
				}
				for _, ask := range event.Asks {
					price, quantity, err := ask.Parse()
					if err != nil {
						return
					}
					d.GetAsks().Update(items_types.NewAsk(items_types.PriceType(price), items_types.QuantityType(quantity)))
				}
				d.LastUpdateID = event.LastUpdateID
			}
		}()
	}
}

func PartialDepthCallBackCreator(
	handlers ...func(d *depths_types.Depths) binance.WsPartialDepthHandler) func(d *depths_types.Depths) binance.WsPartialDepthHandler {
	return func(d *depths_types.Depths) binance.WsPartialDepthHandler {
		var stack []binance.WsPartialDepthHandler
		d.Init()
		standardHandlers := PartialDepthHandlerCreator(d)
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *binance.WsPartialDepthEvent) {
			standardHandlers(event)
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func WsErrorHandlerCreator(handlers ...func(d *depths_types.Depths) binance.ErrHandler) func(*depths_types.Depths) binance.ErrHandler {
	return func(d *depths_types.Depths) binance.ErrHandler {
		var stack []binance.ErrHandler
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(err error) {
			logrus.Errorf("Spot wsErrorHandler error: %v", err)
			d.ResetEvent(err)
			for _, handler := range stack {
				handler(err)
			}
		}
	}
}
