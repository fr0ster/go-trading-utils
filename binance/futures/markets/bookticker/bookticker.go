package bookticker

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetInitCreator(client *futures.Client) func(*booktickers_types.BookTickers) func() (err error) {
	return func(btt *booktickers_types.BookTickers) func() (err error) {
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

func GetStartBookTickerStreamCreator(
	handler futures.WsBookTickerHandler,
	errHandler futures.ErrHandler) func(d *booktickers_types.BookTickers) func() (doneC, stopC chan struct{}, err error) {
	return func(bt *booktickers_types.BookTickers) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsBookTickerServe(bt.GetSymbol(), handler, errHandler)
			return
		}
	}
}
