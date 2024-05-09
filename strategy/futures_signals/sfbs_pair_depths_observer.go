package futures_signals

import (
	"fmt"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_depths "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairPartialDepthsObserver struct {
		client      *futures.Client
		pair        pairs_interfaces.Pairs
		degree      int
		limit       int
		levels      int
		rate        time.Duration
		account     *futures_account.Account
		data        *depth_types.Depth
		depthsEvent chan *futures.WsDepthEvent
		// stream      *futures_streams.PartialDepthServeWithRate
		event     chan bool
		stop      chan os.Signal
		deltaUp   float64
		deltaDown float64
		buyEvent  chan *pair_price_types.PairPrice
		sellEvent chan *pair_price_types.PairPrice
		askUp     chan *pair_price_types.AskBid
		askDown   chan *pair_price_types.AskBid
		bidUp     chan *pair_price_types.AskBid
		bidDown   chan *pair_price_types.AskBid
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
			pp.data = depth_types.New(degree, pp.pair.GetPair())
		}

		// Запускаємо потік для отримання оновлення depths
		logrus.Debugf("Futures, Start stream for %v Klines", pp.pair.GetPair())
		pp.depthsEvent = make(chan *futures.WsDepthEvent, 1)
		wsHandler := func(event *futures.WsDepthEvent) {
			pp.depthsEvent <- event
		}
		futures.WsPartialDepthServeWithRate(pp.pair.GetPair(), pp.levels, pp.rate, wsHandler, utils.HandleErr)
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

func (pp *PairPartialDepthsObserver) StartBuyOrSellSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	if pp.depthsEvent == nil {
		pp.StartStream()
	}
	if pp.buyEvent == nil && pp.sellEvent == nil {
		pp.buyEvent = make(chan *pair_price_types.PairPrice, 1)
		pp.sellEvent = make(chan *pair_price_types.PairPrice, 1)
		buyEvent = pp.buyEvent
		sellEvent = pp.sellEvent
		go func() {
			for {
				if pp.pair.GetMiddlePrice() == 0 {
					continue
				}
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-pp.event: // Чекаємо на спрацювання тригера на зміну bookTicker
					// Кількість базової валюти
					baseBalance, err := GetBaseBalance(pp.account, pp.pair)
					if err != nil {
						logrus.Warnf("Can't get data for analysis: %v", err)
						continue
					}
					pp.pair.SetCurrentBalance(baseBalance)
					pp.pair.SetCurrentPositionBalance(baseBalance * pp.pair.GetLimitOnPosition())
					// Кількість торгової валюти
					targetBalance, err := GetTargetBalance(pp.account, pp.pair)
					if err != nil {
						logrus.Errorf("Can't get %s balance: %v", pp.pair.GetTargetSymbol(), err)
						pp.stop <- os.Interrupt
						return
					}
					// commission := GetCommission(pp.account)
					minAsk := pp.data.GetAsks().Min()
					maxBid := pp.data.GetBids().Max()
					// Ціна купівлі
					ask := minAsk.(*pair_price_types.PairPrice).Price
					// Ціна продажу
					bid := maxBid.(*pair_price_types.PairPrice).Price
					// Верхня межа ціни купівлі
					boundAsk, err := GetAskBound(pp.pair)
					if err != nil {
						logrus.Errorf("Can't get data for analysis: %v", err)
						pp.stop <- os.Interrupt
						return
					}
					// Нижня межа ціни продажу
					boundBid, err := GetBidBound(pp.pair)
					if err != nil {
						logrus.Errorf("Can't get data for analysis: %v", err)
						pp.stop <- os.Interrupt
						return
					}
					// Кількість торгової валюти для продажу
					sellQuantity,
						// Кількість торгової валюти для купівлі
						buyQuantity, err := GetBuyAndSellQuantity(pp.pair, baseBalance, targetBalance, ask, bid)
					if err != nil {
						logrus.Errorf("Can't get data for analysis: %v", err)
						pp.stop <- os.Interrupt
						return
					}

					if buyQuantity == 0 && sellQuantity == 0 {
						logrus.Errorf("We don't have any %s for buy and don't have any %s for sell",
							pp.pair.GetBaseSymbol(), pp.pair.GetTargetSymbol())
						pp.stop <- os.Interrupt
						return
					}
					// Середня ціна купівли цільової валюти більша за верхню межу ціни купівли
					if ask <= boundAsk &&
						targetBalance*ask < pp.pair.GetLimitInputIntoPosition()*baseBalance &&
						targetBalance*ask < pp.pair.GetLimitOutputOfPosition()*baseBalance {
						logrus.Debugf("Middle price %f, Ask %f is lower than high bound price %f, BUY!!!", pp.pair.GetMiddlePrice(), ask, boundAsk)
						buyEvent <- &pair_price_types.PairPrice{
							Price:    ask,
							Quantity: buyQuantity}
						// Середня ціна купівли цільової валюти менша або дорівнює нижній межі ціни продажу
					} else if bid >= boundBid && sellQuantity < targetBalance {
						logrus.Debugf("Middle price %f, Bid %f is higher than low bound price %f, SELL!!!", pp.pair.GetMiddlePrice(), bid, boundBid)
						sellEvent <- &pair_price_types.PairPrice{
							Price:    boundBid,
							Quantity: sellQuantity}
					} else {
						if ask <= boundAsk &&
							(targetBalance*ask > pp.pair.GetLimitInputIntoPosition()*baseBalance ||
								targetBalance*ask > pp.pair.GetLimitOutputOfPosition()*baseBalance) {
							logrus.Debugf("We can't buy %s, because we have more than %f %s",
								pp.pair.GetTargetSymbol(),
								pp.pair.GetLimitInputIntoPosition()*baseBalance,
								pp.pair.GetBaseSymbol())
						} else if bid >= boundBid && sellQuantity >= targetBalance {
							logrus.Debugf("We can't sell %s, because we haven't %s enough for sell, we need %f %s but have %f %s only",
								pp.pair.GetTargetSymbol(),
								pp.pair.GetTargetSymbol(),
								sellQuantity,
								pp.pair.GetTargetSymbol(),
								targetBalance,
								pp.pair.GetTargetSymbol())
						} else if bid < boundBid && ask > boundAsk { // Чекаємо на зміну ціни
							logrus.Debugf("Middle price is %f, bound Bid price %f, bound Ask price %f",
								pp.pair.GetMiddlePrice(), boundBid, boundAsk)
							logrus.Debugf("Wait for buy or sell signal")
							logrus.Debugf("Now ask is %f, bid is %f", ask, bid)
							logrus.Debugf("Waiting for ask decrease to %f or bid increase to %f", boundAsk, boundBid)
						}
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return
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
					pp.stop <- os.Interrupt
					return
				case <-pp.event: // Чекаємо на спрацювання тригера на зміну ціни
					// Ціна купівлі
					ask,
						// Ціна продажу
						bid, err := pp.GetAskBid()
					if err != nil {
						logrus.Errorf("Can't get ask and bid from depth: %v", err)
						pp.stop <- os.Interrupt
						return
					}
					if last_bid == 0 || last_ask == 0 {
						last_bid = bid
						last_ask = ask
					}
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
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.askUp, pp.askDown, pp.bidUp, pp.bidDown
}

func NewPairDepthsObserver(
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	levels int,
	rate time.Duration,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) (pp *PairPartialDepthsObserver, err error) {
	pp = &PairPartialDepthsObserver{
		client:      client,
		pair:        pair,
		account:     nil,
		data:        nil,
		depthsEvent: nil,
		event:       nil,
		stop:        stop,
		degree:      degree,
		limit:       limit,
		levels:      levels,
		rate:        rate,
		deltaUp:     deltaUp,
		deltaDown:   deltaDown,
		askUp:       nil,
		askDown:     nil,
		bidUp:       nil,
		bidDown:     nil,
	}
	pp.account, err = futures_account.New(pp.client, pp.degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
