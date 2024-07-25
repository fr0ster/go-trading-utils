package depth

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	types "github.com/fr0ster/go-trading-utils/types"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/sirupsen/logrus"
)

func InitCreator(limit depth_types.DepthAPILimit, client *futures.Client) func(d *depth_types.Depths) types.InitFunction {
	return func(d *depth_types.Depths) types.InitFunction {
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
	levels depth_types.DepthStreamLevel,
	rate depth_types.DepthStreamRate,
	handlerCreator func(d *depth_types.Depths) futures.WsDepthHandler,
	errHandlerCreator func(d *depth_types.Depths) futures.ErrHandler) func(d *depth_types.Depths) types.StreamFunction {
	return func(d *depth_types.Depths) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsPartialDepthServeWithRate(d.Symbol(), int(levels), time.Duration(rate), handlerCreator(d), errHandlerCreator(d))
			return
		}
	}
}

func eventHandlerCreator(d *depth_types.Depths) futures.WsDepthHandler {
	return func(event *futures.WsDepthEvent) {
		func() {
			d.Lock()         // Locking the depths
			defer d.Unlock() // Unlocking the depths
			if event.LastUpdateID < d.LastUpdateID {
				return
			}
			if event.PrevLastUpdateID != int64(d.LastUpdateID) {
				d.Init()
			} else if event.PrevLastUpdateID == int64(d.LastUpdateID) {
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
	handlers ...func(d *depth_types.Depths) futures.WsDepthHandler) func(d *depth_types.Depths) futures.WsDepthHandler {
	return func(d *depth_types.Depths) futures.WsDepthHandler {
		var stack []futures.WsDepthHandler
		d.Init()
		standardHandler := eventHandlerCreator(d)
		for _, handler := range handlers {
			stack = append(stack, handler(d))
		}
		return func(event *futures.WsDepthEvent) {
			standardHandler(event)
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func WsErrorHandlerCreator(handlers ...func(*depth_types.Depths) futures.ErrHandler) func(*depth_types.Depths) futures.ErrHandler {
	return func(d *depth_types.Depths) futures.ErrHandler {
		var stack []futures.ErrHandler
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
