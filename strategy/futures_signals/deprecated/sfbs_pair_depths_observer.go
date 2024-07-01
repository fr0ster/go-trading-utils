package futures_signals

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_depths "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairPartialDepthsObserver struct {
		client *futures.Client
		// pair         *pairs_types.Pairs
		degree       int
		limit        int
		levels       int
		rate         time.Duration
		account      *futures_account.Account
		exchangeInfo *exchange_info.ExchangeInfo
		data         *depth_types.Depth
		depthsEvent  chan *futures.WsDepthEvent
		event        chan bool
		stop         chan struct{}
		deltaUp      float64
		deltaDown    float64
		askUp        chan *pair_price_types.AskBid
		askDown      chan *pair_price_types.AskBid
		bidUp        chan *pair_price_types.AskBid
		bidDown      chan *pair_price_types.AskBid
		sleepingTime time.Duration
		timeOut      time.Duration
		symbol       *futures.Symbol
	}
)

func (pp *PairPartialDepthsObserver) GetDepths() *depth_types.Depth {
	return pp.data
}

func (pp *PairPartialDepthsObserver) GetStream() chan *futures.WsDepthEvent {
	return pp.depthsEvent
}

func (pp *PairPartialDepthsObserver) StartStream() chan *futures.WsDepthEvent {
	if pp.depthsEvent == nil {
		if pp.data == nil {
			pp.data = depth_types.New(pp.degree, pp.symbol.Symbol)
		}

		ticker := time.NewTicker(pp.timeOut)
		lastResponse := time.Now()
		// Запускаємо потік для отримання оновлення depths
		logrus.Debugf("Futures, Start stream for %v Klines", pp.symbol.Symbol)
		pp.depthsEvent = make(chan *futures.WsDepthEvent, 1)
		wsHandler := func(event *futures.WsDepthEvent) {
			lastResponse = time.Now()
			pp.depthsEvent <- event
		}
		resetEvent := make(chan bool, 1)
		wsErrorHandler := func(err error) {
			resetEvent <- true
		}
		var stopC chan struct{}
		_, stopC, _ = futures.WsPartialDepthServeWithRate(pp.symbol.Symbol, pp.levels, pp.rate, wsHandler, wsErrorHandler)
		go func() {
			for {
				select {
				case <-resetEvent:
					stopC <- struct{}{}
					_, stopC, _ = futures.WsPartialDepthServeWithRate(pp.symbol.Symbol, pp.levels, pp.rate, wsHandler, wsErrorHandler)
				case <-ticker.C:
					if time.Since(lastResponse) > pp.timeOut {
						stopC <- struct{}{}
						_, stopC, _ = futures.WsPartialDepthServeWithRate(pp.symbol.Symbol, pp.levels, pp.rate, wsHandler, wsErrorHandler)
					}
				}
			}
		}()
		futures_depths.Init(pp.data, pp.client, pp.limit)
	}
	return pp.depthsEvent
}

func (pp *PairPartialDepthsObserver) GetAskBid() (bid float64, ask float64, err error) {
	minAsk := pp.data.GetAsks().Min()
	if minAsk == nil {
		err = fmt.Errorf("can't get min ask")
	}
	ask = minAsk.(*pair_price_types.PairPrice).Price
	maxBid := pp.data.GetBids().Max()
	if maxBid == nil {
		err = fmt.Errorf("can't get max bid")
	}
	bid = maxBid.(*pair_price_types.PairPrice).Price
	return
}

func (pp *PairPartialDepthsObserver) StartPriceChangesSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
	if pp.askUp == nil && pp.askDown == nil && pp.bidUp == nil && pp.bidDown == nil {
		pp.askUp = make(chan *pair_price_types.AskBid, 1)
		pp.askDown = make(chan *pair_price_types.AskBid, 1)
		pp.bidUp = make(chan *pair_price_types.AskBid, 1)
		pp.bidDown = make(chan *pair_price_types.AskBid, 1)
		go func() {
			var last_bid, last_ask float64
			for {
				select {
				case <-pp.stop:
					return
				case <-pp.event: // Чекаємо на спрацювання тригера на зміну ціни
					// Ціна купівлі
					ask,
						// Ціна продажу
						bid, err := pp.GetAskBid()
					if err != nil {
						logrus.Errorf("Can't get ask and bid from depth: %v", err)
						close(pp.stop)
						return
					}
					if last_bid == 0 || last_ask == 0 {
						last_bid = bid
						last_ask = ask
					}
					logrus.Debugf("Futures, Ask is %f, Last Ask is %f, Delta Ask is%f%%, Bid is %f, Last Bid is %f, Delta Bid is %f%%",
						ask, last_ask, (ask-last_ask)*100/last_ask, bid, last_bid, (bid-last_bid)*100/last_bid)
					if ask > last_ask*(1+pp.deltaUp) {
						pp.askUp <- &pair_price_types.AskBid{
							Ask: &pair_price_types.PairDelta{Price: ask, Percent: (ask - last_ask) * 100 / last_ask},
							Bid: &pair_price_types.PairDelta{Price: bid, Percent: (bid - last_bid) * 100 / last_bid},
						}
						last_ask = ask
						last_bid = bid
					} else if ask < last_ask*(1-pp.deltaDown) {
						pp.askDown <- &pair_price_types.AskBid{
							Ask: &pair_price_types.PairDelta{Price: ask, Percent: (ask - last_ask) * 100 / last_ask},
							Bid: &pair_price_types.PairDelta{Price: bid, Percent: (bid - last_bid) * 100 / last_bid},
						}
						last_ask = ask
						last_bid = bid
					}
					if bid > last_bid*(1+pp.deltaUp) {
						pp.bidUp <- &pair_price_types.AskBid{
							Ask: &pair_price_types.PairDelta{Price: ask, Percent: (ask - last_ask) * 100 / last_ask},
							Bid: &pair_price_types.PairDelta{Price: bid, Percent: (bid - last_bid) * 100 / last_bid},
						}
						last_ask = ask
						last_bid = bid
					} else if bid < last_bid*(1-pp.deltaDown) {
						pp.bidDown <- &pair_price_types.AskBid{
							Ask: &pair_price_types.PairDelta{Price: ask, Percent: (ask - last_ask) * 100 / last_ask},
							Bid: &pair_price_types.PairDelta{Price: bid, Percent: (bid - last_bid) * 100 / last_bid},
						}
						last_ask = ask
						last_bid = bid
					}
				}
				time.Sleep(pp.sleepingTime)
			}
		}()
	}
	return pp.askUp, pp.askDown, pp.bidUp, pp.bidDown
}

func (pp *PairPartialDepthsObserver) StartUpdateGuard() chan bool {
	if pp.event == nil {
		if pp.depthsEvent == nil {
			pp.StartStream()
		}
		pp.event = futures_handlers.GetDepthsUpdateGuard(pp.data, pp.depthsEvent)
	}
	return pp.event
}

func (pp *PairPartialDepthsObserver) getNotional() (res *futures.MinNotionalFilter, err error) {
	var val *futures.Symbol
	if symbol := pp.exchangeInfo.GetSymbol(&symbol_info.FuturesSymbol{Symbol: pp.symbol.Symbol}); symbol != nil {
		val, err = symbol.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
		if err != nil {
			logrus.Errorf(errorMsg, err)
			return
		}
		res = val.MinNotionalFilter()
	}
	return
}

func (pp *PairPartialDepthsObserver) GetMinQuantity(price float64) float64 {
	notional, err := pp.getNotional()
	if err != nil {
		return 0
	}
	return utils.ConvStrToFloat64(notional.Notional) / price
}

func (pp *PairPartialDepthsObserver) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairPartialDepthsObserver) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func NewPairDepthsObserver(
	client *futures.Client,
	symbol string,
	baseSymbol string,
	targetSymbol string,
	degree int,
	limit int,
	levels int,
	rate time.Duration,
	deltaUp float64,
	deltaDown float64,
	stop chan struct{}) (pp *PairPartialDepthsObserver, err error) {
	pp = &PairPartialDepthsObserver{
		client: client,
		// pair:         pair,
		account:      nil,
		data:         nil,
		depthsEvent:  nil,
		event:        nil,
		stop:         stop,
		degree:       degree,
		limit:        limit,
		levels:       levels,
		rate:         rate,
		deltaUp:      deltaUp,
		deltaDown:    deltaDown,
		askUp:        nil,
		askDown:      nil,
		bidUp:        nil,
		bidDown:      nil,
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}
	pp.account, err = futures_account.New(pp.client, pp.degree, []string{baseSymbol}, []string{targetSymbol})
	if err != nil {
		return
	}
	pp.exchangeInfo = exchange_info.New()
	err = futures_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}
	if symbol := pp.exchangeInfo.GetSymbol(&symbol_info.FuturesSymbol{Symbol: symbol}); symbol != nil {
		pp.symbol, err = symbol.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
		if err != nil {
			logrus.Errorf(errorMsg, err)
			return
		}
	}

	return
}
