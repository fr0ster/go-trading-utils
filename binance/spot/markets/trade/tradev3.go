package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func tradesV3Init(trd []*binance.TradeV3, a *trade_types.TradesV3) (err error) {
	for _, val := range trd {
		a.Update(&trade_types.TradeV3{
			ID:              val.ID,
			Symbol:          val.Symbol,
			OrderID:         val.OrderID,
			OrderListId:     val.OrderListId,
			Price:           val.Price,
			Quantity:        val.Quantity,
			QuoteQuantity:   val.QuoteQuantity,
			Commission:      val.Commission,
			CommissionAsset: val.CommissionAsset,
			Time:            val.Time,
			IsBuyer:         val.IsBuyer,
			IsMaker:         val.IsMaker,
			IsBestMatch:     val.IsBestMatch,
			IsIsolated:      val.IsIsolated,
		})
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
