package depth

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/sirupsen/logrus"
)

func GetterInitCreator(limit depth_types.DepthAPILimit, client *futures.Client) func(d *depth_types.Depths) func() (err error) {
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
	handlerCreator func(d *depth_types.Depths) futures.WsDepthHandler,
	errHandlerCreator func(d *depth_types.Depths) futures.ErrHandler) func(d *depth_types.Depths) func() (doneC, stopC chan struct{}, err error) {
	return func(d *depth_types.Depths) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsPartialDepthServeWithRate(d.Symbol(), int(levels), time.Duration(rate), handlerCreator(d), errHandlerCreator(d))
			return
		}
	}
}

func GetterDepthEventCallBackCreator() func(d *depth_types.Depths) futures.WsDepthHandler {
	return func(d *depth_types.Depths) futures.WsDepthHandler {
		d.Init()
		return func(event *futures.WsDepthEvent) {
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
		}
	}
}

func GetterWsErrorHandlerCreator() func(d *depth_types.Depths) futures.ErrHandler {
	return func(d *depth_types.Depths) futures.ErrHandler {
		return func(err error) {
			logrus.Errorf("Future wsErrorHandler error: %v", err)
			d.ResetEvent(err)
		}
	}
}
