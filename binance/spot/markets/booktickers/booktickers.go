package bookticker

import (
	"context"

	"github.com/adshao/go-binance/v2"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetInitCreator(client *binance.Client) func(*booktickers_types.BookTickers) func() (err error) {
	return func(bt *booktickers_types.BookTickers) func() (err error) {
		return func() (err error) {
			bt.Lock()         // Locking the bookticker
			defer bt.Unlock() // Unlocking the bookticker
			service := client.NewListBookTickersService()
			if bt.GetSymbol() != "" {
				service = service.Symbol(bt.GetSymbol())
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
				bt.Set(bookTicker)
			}
			return nil
		}
	}
}

func GetStartBookTickerStreamCreator(
	handler binance.WsBookTickerHandler,
	errHandler binance.ErrHandler) func(d *booktickers_types.BookTickers) func() (doneC, stopC chan struct{}, err error) {
	return func(bt *booktickers_types.BookTickers) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsBookTickerServe(bt.GetSymbol(), handler, errHandler)
			return
		}
	}
}
