package depth

import (
	"context"

	"github.com/adshao/go-binance/v2"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"

	"github.com/sirupsen/logrus"
)

func GetterInitCreator(limit depth_types.DepthAPILimit, client *binance.Client) func(d *depth_types.Depths) func() (err error) {
	return func(d *depth_types.Depths) func() (err error) {
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

func GetterStartDepthStreamCreator(
	levels depth_types.DepthStreamLevel,
	rate depth_types.DepthStreamRate,
	handlerCreator func(d *depth_types.Depths) binance.WsDepthHandler,
	errHandlerCreator func(d *depth_types.Depths) binance.ErrHandler) func(d *depth_types.Depths) func() (doneC, stopC chan struct{}, err error) {
	return func(d *depth_types.Depths) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsDepthServe(d.Symbol(), handlerCreator(d), errHandlerCreator(d))
			return
		}
	}
}

func standardEventHandlerCreator() func(d *depth_types.Depths) binance.WsDepthHandler {
	return func(d *depth_types.Depths) binance.WsDepthHandler {
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
}

func GetterDepthEventCallBackCreator(
	handlers ...func(d *depth_types.Depths) binance.WsDepthHandler) func(d *depth_types.Depths) binance.WsDepthHandler {
	return func(d *depth_types.Depths) binance.WsDepthHandler {
		d.Init()
		var stack []binance.WsDepthHandler
		d.Init()
		handlers = append(handlers, standardEventHandlerCreator())
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *binance.WsDepthEvent) {
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func GetterStartPartialDepthStreamCreator(
	levels depth_types.DepthStreamLevel,
	rate depth_types.DepthStreamRate,
	handlerCreator func(d *depth_types.Depths) binance.WsPartialDepthHandler,
	errHandlerCreator func(d *depth_types.Depths) binance.ErrHandler) func(d *depth_types.Depths) func() (doneC, stopC chan struct{}, err error) {
	return func(d *depth_types.Depths) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsPartialDepthServe100Ms(d.Symbol(), string(rune(levels)), handlerCreator(d), errHandlerCreator(d))
			return
		}
	}
}

func standardPartialEventHandlerCreator() func(d *depth_types.Depths) binance.WsPartialDepthHandler {
	return func(d *depth_types.Depths) binance.WsPartialDepthHandler {
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
}

func GetterPartialDepthEventCallBackCreator(
	handlers ...func(d *depth_types.Depths) binance.WsPartialDepthHandler) func(d *depth_types.Depths) binance.WsPartialDepthHandler {
	return func(d *depth_types.Depths) binance.WsPartialDepthHandler {
		var stack []binance.WsPartialDepthHandler
		d.Init()
		handlers = append(handlers, standardPartialEventHandlerCreator())
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *binance.WsPartialDepthEvent) {
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func GetterWsErrorHandlerCreator() func(d *depth_types.Depths) binance.ErrHandler {
	return func(d *depth_types.Depths) binance.ErrHandler {
		return func(err error) {
			logrus.Errorf("Spot wsErrorHandler error: %v", err)
			d.ResetEvent(err)
		}
	}
}
