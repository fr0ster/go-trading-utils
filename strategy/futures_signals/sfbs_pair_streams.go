package futures_signals

import (
	"context"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairStreams struct {
		client       *futures.Client
		pair         pairs_interfaces.Pairs
		exchangeInfo *exchange_types.ExchangeInfo
		account      *futures_account.Account

		userDataEvent      chan *futures.WsUserDataEvent
		userDataEvent4AUE  chan *futures.WsUserDataEvent
		accountUpdateEvent chan *futures.WsUserDataEvent

		stop      chan os.Signal
		limitsOut chan bool

		pairInfo   *symbol_types.FuturesSymbol
		orderTypes map[futures.OrderType]bool
		degree     int
		timeOut    time.Duration
	}
)

func (pp *PairStreams) GetExchangeInfo() *exchange_types.ExchangeInfo {
	return pp.exchangeInfo
}

func (pp *PairStreams) GetAccount() *futures_account.Account {
	return pp.account
}

func (pp *PairStreams) GetPairInfo() *symbol_types.FuturesSymbol {
	return pp.pairInfo
}

func (pp *PairStreams) GetOrderTypes() map[futures.OrderType]bool {
	return pp.orderTypes
}

func (pp *PairStreams) GetUserDataEvent() chan *futures.WsUserDataEvent {
	return pp.userDataEvent
}

func (pp *PairStreams) GetAccountUpdateEvent() chan *futures.WsUserDataEvent {
	return pp.accountUpdateEvent
}

func (pp *PairStreams) GetStop() chan os.Signal {
	return pp.stop
}

func (pp *PairStreams) GetLimitsOut() chan bool {
	return pp.limitsOut
}

func (pp *PairStreams) GetDegree() int {
	return pp.degree
}

func (pp *PairStreams) GetTimeOut() time.Duration {
	return pp.timeOut
}

func NewPairStreams(
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	debug bool) (pp *PairStreams, err error) {
	pp = &PairStreams{
		client:       client,
		pair:         pair,
		exchangeInfo: nil,
		account:      nil,

		userDataEvent:      make(chan *futures.WsUserDataEvent),
		userDataEvent4AUE:  make(chan *futures.WsUserDataEvent),
		accountUpdateEvent: nil,

		stop:      make(chan os.Signal, 1),
		limitsOut: make(chan bool, 1),

		pairInfo:   nil,
		orderTypes: make(map[futures.OrderType]bool, 0),
		degree:     3,
		timeOut:    1 * time.Hour,
	}

	// Ініціалізуємо інформацію про біржу
	pp.exchangeInfo = exchange_types.New()
	err = futures_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}

	// Ініціалізуємо інформацію про акаунт
	pp.account, err = futures_account.New(pp.client, pp.degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: pair.GetPair()}).(*symbol_types.FuturesSymbol)

	// Ініціалізуємо типи ордерів які можна використовувати для пари
	pp.orderTypes = make(map[futures.OrderType]bool, 0)
	for _, orderType := range pp.pairInfo.OrderType {
		pp.orderTypes[orderType] = true
	}

	userDataEventStart := func(eventOut chan *futures.WsUserDataEvent) {
		// Ініціалізуємо стріми для відмірювання часу
		ticker := time.NewTicker(pp.timeOut)
		// Ініціалізуємо маркер для останньої відповіді
		lastResponse := time.Now()
		// Отримуємо ключ для прослуховування подій користувача
		listenKey, err := pp.client.NewStartUserStreamService().Do(context.Background())
		if err != nil {
			return
		}
		// Ініціалізуємо канал для відправки подій про необхідність оновлення стріму подій користувача
		resetEvent := make(chan bool, 1)
		// Ініціалізуємо обробник помилок
		wsErrorHandler := func(err error) {
			resetEvent <- true
		}
		// Ініціалізуємо обробник подій
		wsHandler := func(event *futures.WsUserDataEvent) {
			eventOut <- event
		}
		// Запускаємо стрім подій користувача
		var stopC chan struct{}
		_, stopC, err = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
		if err != nil {
			return
		}
		// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
		go func() {
			for {
				select {
				case <-resetEvent:
					// Оновлюємо стан з'єднання для стріму подій користувача з раніше отриманим ключем
					err := pp.client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
					if err != nil {
						// Отримуємо новий ключ для прослуховування подій користувача при втраті з'єднання
						listenKey, err = pp.client.NewStartUserStreamService().Do(context.Background())
						if err != nil {
							return
						}
					}
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо стрім подій користувача
					_, stopC, _ = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
				case <-ticker.C:
					// Оновлюємо стан з'єднання для стріму подій користувача з раніше отриманим ключем
					err := pp.client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
					if err != nil {
						// Отримуємо новий ключ для прослуховування подій користувача при втраті з'єднання
						listenKey, err = pp.client.NewStartUserStreamService().Do(context.Background())
						if err != nil {
							return
						}
					}
					// Перевіряємо чи не вийшли за ліміт часу відповіді
					if time.Since(lastResponse) > pp.timeOut {
						// Зупиняємо стрім подій користувача
						stopC <- struct{}{}
						// Запускаємо стрім подій користувача
						_, stopC, _ = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
					}
				}
			}
		}()
	}
	// Запускаємо стрім подій користувача
	userDataEventStart(pp.userDataEvent)
	userDataEventStart(pp.userDataEvent4AUE)

	// Запускаємо стрім для відслідковування зміни статусу акаунта
	pp.accountUpdateEvent = futures_handlers.GetAccountInfoGuard(pp.account, pp.userDataEvent4AUE)

	return
}
