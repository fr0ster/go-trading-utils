package bookticker

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

func InitCreator(client *binance.Client) func(*booktickers_types.BookTickers) types.InitFunction {
	return func(bt *booktickers_types.BookTickers) types.InitFunction {
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

func BookTickerStreamCreator(
	handler func(*booktickers_types.BookTickers) binance.WsBookTickerHandler,
	errHandler func(*booktickers_types.BookTickers) binance.ErrHandler) func(*booktickers_types.BookTickers) types.StreamFunction {
	return func(bt *booktickers_types.BookTickers) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsBookTickerServe(bt.GetSymbol(), handler(bt), errHandler(bt))
			return
		}
	}
}

func eventHandlerCreator(bt *booktickers_types.BookTickers) binance.WsBookTickerHandler {
	return func(event *binance.WsBookTickerEvent) {
		func() {
			bt.Lock()         // Locking the depths
			defer bt.Unlock() // Unlocking the depths
			if btt := bt.Get(event.Symbol); btt != nil {
				btt.SetAskPrice(depths_types.PriceType(utils.ConvStrToFloat64(event.BestAskPrice)))
				btt.SetAskQuantity(depths_types.QuantityType(utils.ConvStrToFloat64(event.BestAskQty)))
				btt.SetBidPrice(depths_types.PriceType(utils.ConvStrToFloat64(event.BestBidPrice)))
				btt.SetBidQuantity(depths_types.QuantityType(utils.ConvStrToFloat64(event.BestBidQty)))
				btt.SetUpdateID(event.UpdateID)
				bt.Set(btt)
			}
		}()
	}
}

func CallBackCreator(
	handlers ...func(*booktickers_types.BookTickers) binance.WsBookTickerHandler) func(*booktickers_types.BookTickers) binance.WsBookTickerHandler {
	return func(bt *booktickers_types.BookTickers) binance.WsBookTickerHandler {
		var stack []binance.WsBookTickerHandler
		standardHandlers := eventHandlerCreator(bt)
		for _, handler := range handlers {
			stack = append(stack, handler(bt))
		}
		return func(event *binance.WsBookTickerEvent) {
			standardHandlers(event)
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}

func WsErrorHandlerCreator() func(*booktickers_types.BookTickers) binance.ErrHandler {
	return func(btt *booktickers_types.BookTickers) binance.ErrHandler {
		return func(err error) {
			logrus.Errorf("Future BookTickers error: %v", err)
			btt.ResetEvent(err)
		}
	}
}
