package futures_signals

import (
	"context"
	_ "net/http/pprof"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func StartPairStreams(
	symbol string,
	bookTicker *bookTicker_types.BookTickerBTree,
	depth *depth_types.Depth) (
	depthEvent chan bool,
	bookTickerEvent chan bool) {
	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := futures_streams.NewBookTickerStream(symbol, 1)
	bookTickerStream.Start()

	bookTickerEvent = futures_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	// Запускаємо потік для отримання оновлення стакана
	depthStream := futures_streams.NewDiffDepthStream(symbol, 1)
	depthStream.Start()

	depthEvent = futures_handlers.GetDepthsUpdateGuard(depth, depthStream.DataChannel)

	return
}

func StartGlobalStreams(
	client *futures.Client,
	stop chan os.Signal,
	account *futures_account.Account) (
	userDataStream4Account *futures_streams.UserDataStream,
	accountUpdateEvent chan *futures.WsUserDataEvent,
	userDataStream4Order *futures_streams.UserDataStream,
	orderStatusEvent chan *futures.WsUserDataEvent) {
	// Запускаємо потік для отримання wsUserDataEvent
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		logrus.Errorf(errorMsg, err)
		stop <- os.Interrupt
		return
	}
	userDataStream4Order = futures_streams.NewUserDataStream(listenKey, 1)
	userDataStream4Order.Start()

	orderStatuses := []futures.OrderStatusType{
		futures.OrderStatusTypeFilled,
		futures.OrderStatusTypePartiallyFilled,
	}

	orderStatusEvent = futures_handlers.GetChangingOfOrdersGuard(
		userDataStream4Order.GetDataChannel(),
		orderStatuses)

	userDataStream4Account = futures_streams.NewUserDataStream(listenKey, 1)
	userDataStream4Account.Start()

	// Запускаємо потік для отримання оновлення аккаунту
	accountUpdateEvent = futures_handlers.GetAccountInfoGuard(account, userDataStream4Account.GetDataChannel())

	return
}
