package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func tradesV3Init(trd []*binance.TradeV3, a *trade_types.TradesV3) (err error) {
	a.Lock()         // Locking the trades
	defer a.Unlock() // Unlocking the trades
	for _, val := range trd {
		tradeV3, err := trade_types.Binance2TradesV3(val)
		if err != nil {
			return err
		}
		a.Update(tradeV3)
	}
	return nil
}

func ListTradesInit(a *trade_types.TradesV3, client *binance.Client, symbolname string, limit int) (err error) {
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

func ListMarginTradesInit(a *trade_types.TradesV3, client *binance.Client, symbolname string, limit int) (err error) {
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
