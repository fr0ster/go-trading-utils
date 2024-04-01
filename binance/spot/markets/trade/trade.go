package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func tradesInit(trd []*binance.Trade, a *trade_types.Trades) (err error) {
	a.Lock()         // Locking the trades
	defer a.Unlock() // Unlocking the trades
	for _, val := range trd {
		trade, err := trade_types.Binance2Trades(val)
		if err != nil {
			return err
		}
		a.Update(trade)
	}
	return nil
}

func HistoricalTradesInit(a *trade_types.Trades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewHistoricalTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesInit(res, a)
}

func RecentTradesInit(a *trade_types.Trades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(apt_key, secret_key)
	res, err :=
		client.NewRecentTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesInit(res, a)
}
