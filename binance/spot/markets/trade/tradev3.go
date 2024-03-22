package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func tradesV3Init(res []*binance.TradeV3, a *trade_types.TradesV3) (err error) {
	for _, val := range res {
		a.Update(trade_types.TradeV3(*val))
	}
	return nil
}

func ListTradesInit(a *trade_types.TradesV3, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewListTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesV3Init(res, a)
}

func ListMarginTradesInit(a *trade_types.TradesV3, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewListMarginTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesV3Init(res, a)
}
