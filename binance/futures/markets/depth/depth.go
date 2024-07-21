package depth

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/sirupsen/logrus"
)

func Init(d *depth_types.Depths, limit depths_types.DepthAPILimit, client *futures.Client) (err error) {
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

func GetStartDepthStream(
	d *depth_types.Depths,
	levels depths_types.DepthStreamLevel,
	rate depths_types.DepthStreamRate,
	handler futures.WsDepthHandler,
	errHandler futures.ErrHandler) func() (
	doneC,
	stopC chan struct{},
	err error) {
	return func() (doneC, stopC chan struct{}, err error) {
		// Запускаємо стрім подій користувача
		doneC, stopC, err = futures.WsPartialDepthServeWithRate(d.Symbol(), int(levels), time.Duration(rate), handler, errHandler)
		return
	}
}

func GetDepthEventCallBack(d *depth_types.Depths) futures.WsDepthHandler {
	d.Init(d)
	return func(event *futures.WsDepthEvent) {
		d.Lock()         // Locking the depths
		defer d.Unlock() // Unlocking the depths
		if event.LastUpdateID < d.LastUpdateID {
			return
		}
		if event.PrevLastUpdateID != int64(d.LastUpdateID) {
			d.Init(d)
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

func GetWsErrorHandler(d *depth_types.Depths) futures.ErrHandler {
	return func(err error) {
		logrus.Errorf("Future wsErrorHandler error: %v", err)
		d.ResetEvent(err)
	}
}
