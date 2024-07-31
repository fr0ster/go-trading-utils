package booktickers

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

func InitCreator(client *futures.Client) func(*booktickers_types.BookTickers) types.InitFunction {
	return func(btt *booktickers_types.BookTickers) types.InitFunction {
		return func() (err error) {
			btt.Lock()         // Locking the bookticker
			defer btt.Unlock() // Unlocking the bookticker
			service := client.NewListBookTickersService()
			if btt.GetSymbol() != "" {
				service = service.Symbol(string(btt.GetSymbol()))
			}
			bookTickerList, err := service.Do(context.Background())
			if err != nil {
				return
			}
			for _, bookTicker := range bookTickerList {
				bookTicker := bookticker_types.New(
					bookTicker.Symbol,
					depths_types.PriceType(utils.ConvStrToFloat64(bookTicker.BidPrice)),
					depths_types.QuantityType(utils.ConvStrToFloat64(bookTicker.BidQuantity)),
					depths_types.PriceType(utils.ConvStrToFloat64(bookTicker.AskPrice)),
					depths_types.QuantityType(utils.ConvStrToFloat64(bookTicker.AskQuantity)))
				btt.Set(bookTicker)
			}
			return nil
		}
	}
}

func BookTickerStreamCreator(
	handler futures.WsBookTickerHandler,
	errHandler futures.ErrHandler) func(d *booktickers_types.BookTickers) types.StreamFunction {
	return func(bt *booktickers_types.BookTickers) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsBookTickerServe(bt.GetSymbol(), handler, errHandler)
			return
		}
	}
}

func WsErrorHandlerCreator() func(bt *booktickers_types.BookTickers) futures.ErrHandler {
	return func(bt *booktickers_types.BookTickers) futures.ErrHandler {
		return func(err error) {
			logrus.Errorf("Future BookTickers error: %v", err)
			bt.ResetEvent(err)
		}
	}
}
