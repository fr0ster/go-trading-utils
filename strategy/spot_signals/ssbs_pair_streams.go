package spot_signals

import (
	"context"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	PairStreams struct {
		client       *binance.Client
		pair         *pairs_types.Pairs
		exchangeInfo *exchange_types.ExchangeInfo
		account      *spot_account.Account

		userDataEvent      chan *binance.WsUserDataEvent
		userDataEvent4AUE  chan *binance.WsUserDataEvent
		accountUpdateEvent chan *binance.WsUserDataEvent

		stop      chan os.Signal
		limitsOut chan bool

		pairInfo   *symbol_types.SpotSymbol
		orderTypes map[string]bool
		degree     int
		timeOut    time.Duration
	}
)

func (pp *PairStreams) GetExchangeInfo() *exchange_types.ExchangeInfo {
	return pp.exchangeInfo
}

func (pp *PairStreams) GetAccount() *spot_account.Account {
	return pp.account
}

func (pp *PairStreams) GetPairInfo() *symbol_types.SpotSymbol {
	return pp.pairInfo
}

func (pp *PairStreams) GetOrderTypes() map[string]bool {
	return pp.orderTypes
}

func (pp *PairStreams) GetUserDataEvent() chan *binance.WsUserDataEvent {
	return pp.userDataEvent
}

func (pp *PairStreams) GetUserDataEvent4AUE() chan *binance.WsUserDataEvent {
	return pp.userDataEvent4AUE
}

func (pp *PairStreams) GetAccountUpdateEvent() chan *binance.WsUserDataEvent {
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
	client *binance.Client,
	pair *pairs_types.Pairs,
	debug bool) (pp *PairStreams, err error) {
	pp = &PairStreams{
		client:  client,
		pair:    pair,
		account: nil,

		userDataEvent:      make(chan *binance.WsUserDataEvent),
		userDataEvent4AUE:  make(chan *binance.WsUserDataEvent),
		accountUpdateEvent: nil,

		stop:      make(chan os.Signal, 1),
		limitsOut: make(chan bool, 1),

		pairInfo:   nil,
		orderTypes: make(map[string]bool, 0),
		degree:     3,
		timeOut:    1 * time.Hour,
	}

	// Ініціалізуємо інформацію про біржу
	pp.exchangeInfo = exchange_types.New()
	err = spot_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}

	// Ініціалізуємо інформацію про акаунт
	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.SpotSymbol{Symbol: pair.GetPair()}).(*symbol_types.SpotSymbol)

	// Ініціалізуємо типи ордерів які можна використовувати для пари
	for _, orderType := range pp.pairInfo.OrderTypes {
		pp.orderTypes[orderType] = true
	}

	userDataEventStart := func(eventOut chan *binance.WsUserDataEvent) {
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
		wsHandler := func(event *binance.WsUserDataEvent) {
			eventOut <- event
		}
		// Запускаємо стрім подій користувача
		var stopC chan struct{}
		_, stopC, err = binance.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
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
					_, stopC, _ = binance.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
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
						_, stopC, _ = binance.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
					}
				}
			}
		}()
	}
	// Запускаємо стрім подій користувача
	userDataEventStart(pp.userDataEvent)
	userDataEventStart(pp.userDataEvent4AUE)

	// Запускаємо стрім для відслідковування оновлення акаунта
	pp.accountUpdateEvent = spot_handlers.GetAccountInfoGuard(pp.account, pp.userDataEvent4AUE)

	return
}
