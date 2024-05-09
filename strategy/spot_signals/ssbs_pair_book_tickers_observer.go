package spot_signals

import (
	"fmt"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_book_ticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairBookTickersObserver struct {
		client          *binance.Client
		pair            pairs_interfaces.Pairs
		degree          int
		limit           int
		account         *spot_account.Account
		data            *book_ticker_types.BookTickers
		bookTickerEvent chan *binance.WsBookTickerEvent
		event           chan bool
		stop            chan os.Signal
		deltaUp         float64
		deltaDown       float64
		buyEvent        chan *pair_price_types.PairPrice
		sellEvent       chan *pair_price_types.PairPrice
		askUp           chan *pair_price_types.AskBid
		askDown         chan *pair_price_types.AskBid
		bidUp           chan *pair_price_types.AskBid
		bidDown         chan *pair_price_types.AskBid
	}
)

func (pp *PairBookTickersObserver) GetBookTickers() *book_ticker_types.BookTicker {
	btk := pp.data.Get(pp.pair.GetPair())
	if btk == nil {
		return nil
	}
	return btk.(*book_ticker_types.BookTicker)
}

func (pp *PairBookTickersObserver) GetStream() chan *binance.WsBookTickerEvent {
	return pp.bookTickerEvent
}

func (pp *PairBookTickersObserver) StartStream() chan *binance.WsBookTickerEvent {
	if pp.bookTickerEvent == nil {
		if pp.data == nil {
			pp.data = book_ticker_types.New(degree)
		}

		// Запускаємо потік для отримання оновлення bookTickers
		pp.bookTickerEvent = make(chan *binance.WsBookTickerEvent, 1)
		logrus.Debugf("Spot, Start stream for %v Klines", pp.pair.GetPair())
		wsHandler := func(event *binance.WsBookTickerEvent) {
			pp.bookTickerEvent <- event
		}
		binance.WsBookTickerServe(pp.pair.GetPair(), wsHandler, utils.HandleErr)
		spot_book_ticker.Init(pp.data, pp.pair.GetPair(), pp.client)
	}
	return pp.bookTickerEvent
}

func (pp *PairBookTickersObserver) GetAskBid() (bid float64, ask float64, err error) {
	btk := pp.data.Get(pp.pair.GetPair())
	if btk == nil {
		err = fmt.Errorf("can't get bookTicker for %s", pp.pair.GetPair())
		return
	}
	ask = btk.(*book_ticker_types.BookTicker).AskPrice
	bid = btk.(*book_ticker_types.BookTicker).BidPrice
	return
}

func (pp *PairBookTickersObserver) StartBuyOrSellSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	if pp.bookTickerEvent == nil {
		pp.StartStream()
	}
	if pp.event == nil {
		pp.StartUpdateGuard()
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
					commission := GetCommission(pp.account)
					bookTicker := pp.data.Get(pp.pair.GetPair())
					if bookTicker == nil {
						logrus.Errorf("Can't get bookTicker for %s", pp.pair.GetPair())
						pp.stop <- os.Interrupt
						return
					}
					// Ціна купівлі
					ask := bookTicker.(*book_ticker_types.BookTicker).AskPrice
					// Ціна продажу
					bid := bookTicker.(*book_ticker_types.BookTicker).AskPrice
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
						buyQuantity, err := GetBuyAndSellQuantity(pp.pair, baseBalance, targetBalance, commission, commission, ask, bid)
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

func (pp *PairBookTickersObserver) StartPriceChangesSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
	bookTicker := pp.data.Get(pp.pair.GetPair())
	if pp.data == nil {
		pp.data = book_ticker_types.New(degree)
	}
	if bookTicker == nil {
		logrus.Errorf("Can't get bookTicker for %s when read for last price, spot strategy", pp.pair.GetPair())
		pp.stop <- os.Interrupt
		return
	}
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
					bookTicker := pp.data.Get(pp.pair.GetPair())
					if bookTicker == nil {
						logrus.Errorf("Can't get bookTicker for %s", pp.pair.GetPair())
						pp.stop <- os.Interrupt
						return
					}
					// Ціна купівлі
					ask := bookTicker.(*book_ticker_types.BookTicker).AskPrice
					// Ціна продажу
					bid := bookTicker.(*book_ticker_types.BookTicker).BidPrice
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

func (pp *PairBookTickersObserver) StartUpdateGuard() chan bool {
	if pp.event == nil {
		if pp.bookTickerEvent == nil {
			pp.StartStream()
		}
		pp.event = spot_handlers.GetBookTickersUpdateGuard(pp.data, pp.bookTickerEvent)
	}
	return pp.event
}

func NewPairBookTickersObserver(
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) (pp *PairBookTickersObserver, err error) {
	pp = &PairBookTickersObserver{
		client:          client,
		pair:            pair,
		account:         nil,
		data:            nil,
		bookTickerEvent: nil,
		event:           nil,
		stop:            stop,
		degree:          degree,
		limit:           limit,
		deltaUp:         deltaUp,
		deltaDown:       deltaDown,
		askUp:           nil,
		askDown:         nil,
		bidUp:           nil,
		bidDown:         nil,
	}
	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
