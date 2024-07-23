package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/tradeV3"
	"github.com/sirupsen/logrus"
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

func ListTradesInitCreator(client *binance.Client, limit int) func(a *trade_types.TradesV3) types.InitFunction {
	return func(a *trade_types.TradesV3) types.InitFunction {
		return func() (err error) {
			res, err :=
				client.NewListTradesService().
					Symbol(string(a.GetSymbolname())).
					Limit(limit).
					Do(context.Background())
			if err != nil {
				return err
			}
			return tradesV3Init(res, a)
		}
	}
}

func ListMarginTradesInitCreator(client *binance.Client, limit int) func(a *trade_types.TradesV3) types.InitFunction {
	return func(a *trade_types.TradesV3) types.InitFunction {
		return func() (err error) {
			res, err :=
				client.NewListMarginTradesService().
					Symbol(string(a.GetSymbolname())).
					Limit(limit).
					Do(context.Background())
			if err != nil {
				return err
			}
			return tradesV3Init(res, a)
		}
	}
}

func WsErrorHandlerCreator() func(*trade_types.TradesV3) binance.ErrHandler {
	return func(trade *trade_types.TradesV3) binance.ErrHandler {
		return func(err error) {
			logrus.Errorf("Spot wsErrorHandler error: %v", err)
			trade.ResetEvent(err)
		}
	}
}
