package spot_signals

import (
	"context"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairStreams struct {
		client       *binance.Client
		pair         pairs_interfaces.Pairs
		exchangeInfo *exchange_types.ExchangeInfo
		account      *spot_account.Account

		userDataEvent    chan *binance.WsUserDataEvent
		orderStatusEvent chan *binance.WsUserDataEvent

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

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

func (pp *PairStreams) GetMinuteOrderLimit() *exchange_types.RateLimits {
	return pp.minuteOrderLimit
}

func (pp *PairStreams) GetDayOrderLimit() *exchange_types.RateLimits {
	return pp.dayOrderLimit
}

func (pp *PairStreams) GetMinuteRawRequestLimit() *exchange_types.RateLimits {
	return pp.minuteRawRequestLimit
}

func (pp *PairStreams) GetOrderStatusEvent() chan *binance.WsUserDataEvent {
	return pp.orderStatusEvent
}

func (pp *PairStreams) GetUserDataEvent() chan *binance.WsUserDataEvent {
	return pp.userDataEvent
}

func (pp *PairStreams) GetStop() chan os.Signal {
	return pp.stop
}

func (pp *PairStreams) GetLimitsOut() chan bool {
	return pp.limitsOut
}

func (pp *PairStreams) GetUpdateTime() time.Duration {
	return pp.updateTime
}

func (pp *PairStreams) GetDegree() int {
	return pp.degree
}

func (pp *PairStreams) GetTimeOut() time.Duration {
	return pp.timeOut
}

func (pp *PairStreams) LimitUpdaterStream() {
	go func() {
		for {
			select {
			case <-time.After(pp.updateTime):
				pp.updateTime,
					pp.minuteOrderLimit,
					pp.dayOrderLimit,
					pp.minuteRawRequestLimit = LimitRead(pp.degree, []string{pp.pair.GetPair()}, pp.client)
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			}
		}
	}()

	// Перевіряємо чи не вийшли за ліміти на запити та ордери
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
			case <-pp.limitsOut:
				pp.stop <- os.Interrupt
				return
			default:
			}
			time.Sleep(pp.updateTime)
		}
	}()
}

func NewPairStreams(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	debug bool) (pp *PairStreams, err error) {
	pp = &PairStreams{
		client:    client,
		pair:      pair,
		account:   nil,
		stop:      make(chan os.Signal, 1),
		limitsOut: make(chan bool, 1),
		pairInfo:  nil,

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		degree:  3,
		timeOut: 1 * time.Hour,
	}

	// Перевіряємо ліміти на ордери та запити
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit =
		LimitRead(degree, []string{pp.pair.GetPair()}, client)

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
	pp.orderTypes = make(map[string]bool, 0)
	for _, orderType := range pp.pairInfo.OrderTypes {
		pp.orderTypes[orderType] = true
	}

	// Ініціалізуємо стріми для оновлення лімітів на ордери та запити
	pp.LimitUpdaterStream()

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
		pp.userDataEvent <- event
	}
	// Ініціалізуємо канал подій користувача
	pp.userDataEvent = make(chan *binance.WsUserDataEvent)
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

	// Визначаємо статуси ордерів які нас цікавлять
	orderStatuses := []binance.OrderStatusType{
		binance.OrderStatusTypeFilled,
		binance.OrderStatusTypePartiallyFilled,
	}
	// Запускаємо стрім для відслідковування зміни статусу ордерів які нас цікавлять
	pp.orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(pp.userDataEvent, orderStatuses)

	return
}
