package spot_signals

import (
	"context"
	_ "net/http/pprof"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func StartPairStreams(
	symbol string,
	bookTicker *book_ticker_types.BookTickers,
	depth *depth_types.Depth) (
	depthEvent chan bool,
	bookTickerEvent chan bool) {
	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(symbol, 1)
	bookTickerStream.Start()

	bookTickerEvent = spot_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	// Запускаємо потік для отримання оновлення стакана
	depthStream := spot_streams.NewDepthStream(symbol, true, 1)
	depthStream.Start()

	depthEvent = spot_handlers.GetDepthsUpdateGuard(depth, depthStream.DataChannel)

	return
}

func StartGlobalStreams(
	client *binance.Client,
	stop chan os.Signal,
	account *spot_account.Account) (
	userDataStream4Account *spot_streams.UserDataStream,
	accountUpdateEvent chan *binance.WsUserDataEvent,
	userDataStream4Order *spot_streams.UserDataStream,
	orderStatusEvent chan *binance.WsUserDataEvent) {
	// Запускаємо потік для отримання wsUserDataEvent
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		logrus.Errorf(errorMsg, err)
		stop <- os.Interrupt
		return
	}
	userDataStream4Order = spot_streams.NewUserDataStream(listenKey, 1)
	userDataStream4Order.Start()

	orderStatuses := []binance.OrderStatusType{
		binance.OrderStatusTypeFilled,
		binance.OrderStatusTypePartiallyFilled,
	}

	orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(
		userDataStream4Order.GetDataChannel(),
		orderStatuses)

	userDataStream4Account = spot_streams.NewUserDataStream(listenKey, 1)
	userDataStream4Account.Start()

	// Запускаємо потік для отримання оновлення аккаунту
	accountUpdateEvent = spot_handlers.GetAccountInfoGuard(account, userDataStream4Account.GetDataChannel())

	return
}
