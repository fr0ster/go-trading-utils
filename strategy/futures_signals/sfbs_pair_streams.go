package futures_signals

import (
	"context"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	PairStreams struct {
		client       *futures.Client
		pair         *pairs_types.Pairs
		exchangeInfo *exchange_types.ExchangeInfo
		account      *futures_account.Account

		userDataEvent     chan *futures.WsUserDataEvent
		userDataEvent4AUE chan *futures.WsUserDataEvent
		// userDataEvent4OTU        chan *futures.WsUserDataEvent
		// userDataEvent4ACU        chan *futures.WsUserDataEvent
		accountUpdateEvent       chan futures.WsAccountUpdate
		orderTradeUpdateEvent    chan futures.WsOrderTradeUpdate
		accountConfigUpdateEvent chan futures.WsAccountConfigUpdate

		stop chan struct{}

		pairInfo     *symbol_types.FuturesSymbol
		orderTypes   map[futures.OrderType]bool
		degree       int
		timeOut      time.Duration
		eventTimeOut time.Duration
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

func (pp *PairStreams) GetAccountUpdateEvent() chan futures.WsAccountUpdate {
	return pp.accountUpdateEvent
}

func (pp *PairStreams) GetOrderTradeUpdateEvent() chan futures.WsOrderTradeUpdate {
	return pp.orderTradeUpdateEvent
}

func (pp *PairStreams) GetAccountConfigUpdateEvent() chan futures.WsAccountConfigUpdate {
	return pp.accountConfigUpdateEvent
}

func (pp *PairStreams) GetStop() chan struct{} {
	return pp.stop
}

func (pp *PairStreams) GetDegree() int {
	return pp.degree
}

func (pp *PairStreams) GetTimeOut() time.Duration {
	return pp.timeOut
}

func (pp *PairStreams) GetPositionRisk() (risks *futures.PositionRisk, err error) {
	risks, err = pp.GetAccount().GetPositionRisk(pp.pair.GetPair())
	return
}

func (pp *PairStreams) GetLiquidationDistance(price float64) (distance float64) {
	risk, _ := pp.GetPositionRisk()
	return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
}

func NewPairStreams(
	client *futures.Client,
	pair *pairs_types.Pairs,
	stop chan struct{},
	debug bool) (pp *PairStreams, err error) {
	pp = &PairStreams{
		client:       client,
		pair:         pair,
		exchangeInfo: nil,
		account:      nil,

		userDataEvent:      make(chan *futures.WsUserDataEvent),
		userDataEvent4AUE:  make(chan *futures.WsUserDataEvent),
		accountUpdateEvent: nil,

		stop: stop,

		pairInfo:     nil,
		orderTypes:   make(map[futures.OrderType]bool, 0),
		degree:       3,
		timeOut:      1 * time.Hour,
		eventTimeOut: 100 * time.Millisecond,
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

	userDataEventStart := func(eventOut chan *futures.WsUserDataEvent, eventType ...futures.UserDataEventType) {
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
			if len(eventType) != 1 || event.Event == eventType[0] {
				eventOut <- event
			}
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
	// userDataEventStart(pp.userDataEvent4AUE, futures.UserDataEventTypeAccountUpdate)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-pp.stop:
	// 			return
	// 		case event := <-pp.userDataEvent4AUE:
	// 			logrus.Debugf("Futures %s: AUE event %v", pp.pair.GetPair(), event)
	// 			pp.accountUpdateEvent <- event.AccountUpdate
	// 		}
	// 		time.Sleep(pp.eventTimeOut)
	// 	}
	// }()
	// userDataEventStart(pp.userDataEvent4OTU, futures.UserDataEventTypeOrderTradeUpdate)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-pp.stop:
	// 			return
	// 		case event := <-pp.userDataEvent4OTU:
	// 			logrus.Debugf("Futures %s: OTU event %v", pp.pair.GetPair(), event)
	// 			pp.orderTradeUpdateEvent <- event.OrderTradeUpdate
	// 		}
	// 		time.Sleep(pp.eventTimeOut)
	// 	}
	// }()
	// userDataEventStart(pp.userDataEvent4ACU, futures.UserDataEventTypeAccountConfigUpdate)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-pp.stop:
	// 			return
	// 		case event := <-pp.userDataEvent4ACU:
	// 			logrus.Debugf("Futures %s: ACU event %v", pp.pair.GetPair(), event)
	// 			pp.accountConfigUpdateEvent <- event.AccountConfigUpdate
	// 		}
	// 		time.Sleep(pp.eventTimeOut)
	// 	}
	// }()
	return
}
