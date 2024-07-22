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
	handler func(*booktickers_types.BookTickers) binance.WsBookTickerHandler,
	errHandler func(*booktickers_types.BookTickers) binance.ErrHandler) func(*booktickers_types.BookTickers) func() (doneC, stopC chan struct{}, err error) {
	return func(bt *booktickers_types.BookTickers) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsBookTickerServe(bt.GetSymbol(), handler(bt), errHandler(bt))
			return
		}
	}
}

func standardEventHandlerCreator(bt *booktickers_types.BookTickers) binance.WsBookTickerHandler {
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

func StandardEventCallBackCreator(
	handlers ...func(*booktickers_types.BookTickers) binance.WsBookTickerHandler) func(*booktickers_types.BookTickers) binance.WsBookTickerHandler {
	return func(bt *booktickers_types.BookTickers) binance.WsBookTickerHandler {
		var stack []binance.WsBookTickerHandler
		standardHandlers := standardEventHandlerCreator(bt)
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
