package depth

import (
	"context"

	"github.com/adshao/go-binance/v2"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/sirupsen/logrus"
)

func Init(d *depth_types.Depths, limit depths_types.DepthAPILimit, client *binance.Client) (err error) {
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
	handler binance.WsDepthHandler,
	errHandler binance.ErrHandler) func() (
	doneC,
	stopC chan struct{},
	err error) {
	return func() (doneC, stopC chan struct{}, err error) {
		// Запускаємо стрім подій користувача
		doneC, stopC, err = binance.WsDepthServe100Ms(d.Symbol(), handler, errHandler)
		return
	}
}

func GetStartPartialDepthStream(
	d *depth_types.Depths,
	levels depths_types.DepthStreamLevel,
	rate depths_types.DepthStreamRate,
	handler binance.WsPartialDepthHandler,
	errHandler binance.ErrHandler) func() (
	doneC,
	stopC chan struct{},
	err error) {
	return func() (doneC, stopC chan struct{}, err error) {
		// Запускаємо стрім подій користувача
		doneC, stopC, err = binance.WsPartialDepthServe100Ms(d.Symbol(), string(rune(levels)), handler, errHandler)
		return
	}
}

func GetWsErrorHandler(d *depth_types.Depths) binance.ErrHandler {
	return func(err error) {
		logrus.Errorf("Spot wsErrorHandler error: %v", err)
		d.ResetEvent(err)
	}
}
