package futures_signals

import (
	"fmt"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_book_ticker "github.com/fr0ster/go-trading-utils/binance/futures/markets/bookticker"
	futures_depths "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	futures_price "github.com/fr0ster/go-trading-utils/binance/futures/markets/price"

	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

	utils "github.com/fr0ster/go-trading-utils/utils"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
	price_types "github.com/fr0ster/go-trading-utils/types/price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairObserver struct {
		client             *futures.Client
		pair               pairs_interfaces.Pairs
		account            *futures_account.Account
		bookTickers        *book_ticker_types.BookTickers
		bookTickerStream   *futures_streams.BookTickerStream
		degree             int
		limit              int
		depths             *depth_types.Depth
		depthsStream       *futures_streams.PartialDepthServeWithRate
		bookTickerEvent    chan bool
		depthEvent         chan bool
		collectionOutEvent chan bool
		positionOutEvent   chan bool
		priceChanges       chan *pair_price_types.PairDelta
		riskEvent          chan bool
		priceUp            chan bool
		priceDown          chan bool
		stop               chan os.Signal
		deltaUp            float64
		deltaDown          float64
		askUp              chan *pair_price_types.AskBid
		askDown            chan *pair_price_types.AskBid
		bidUp              chan *pair_price_types.AskBid
		bidDown            chan *pair_price_types.AskBid
	}
)

func (pp *PairObserver) GetBookTicker() *book_ticker_types.BookTicker {
	btk := pp.bookTickers.Get(pp.pair.GetPair())
	if btk == nil {
		return nil
	}
	return btk.(*book_ticker_types.BookTicker)
}

func (pp *PairObserver) GetBookTickerStream() *futures_streams.BookTickerStream {
	return pp.bookTickerStream
}

func (pp *PairObserver) GetDepth() *depth_types.Depth {
	return pp.depths
}

func (pp *PairObserver) GetDepthStream() *futures_streams.PartialDepthServeWithRate {
	return pp.depthsStream
}

func (pp *PairObserver) GetBookTickerAskBid() (bid float64, ask float64, err error) {
	btk := pp.bookTickers.Get(pp.pair.GetPair())
	if btk == nil {
		err = fmt.Errorf("can't get bookTicker for %s", pp.pair.GetPair())
		return
	}
	ask = btk.(*book_ticker_types.BookTicker).AskPrice
	bid = btk.(*book_ticker_types.BookTicker).BidPrice
	return
}

func (pp *PairObserver) GetDepthAskBid() (bid float64, ask float64, err error) {
	minAsk := pp.depths.GetAsks().Min()
	if minAsk == nil {
		err = fmt.Errorf("can't get min ask")
	}
	ask = minAsk.(*pair_price_types.PairPrice).Price
	maxBid := pp.depths.GetBids().Max()
	if maxBid == nil {
		err = fmt.Errorf("can't get max bid")
	}
	bid = maxBid.(*pair_price_types.PairPrice).Price
	return
}

func (pp *PairObserver) StartBookTickerStream() *futures_streams.BookTickerStream {
	if pp.bookTickerStream != nil {
		if pp.bookTickers == nil {
			pp.bookTickers = book_ticker_types.New(degree)
		}

		// Запускаємо потік для отримання оновлення bookTickers
		pp.bookTickerStream = futures_streams.NewBookTickerStream(pp.pair.GetPair(), 1)
		pp.bookTickerStream.Start()
		futures_book_ticker.Init(pp.bookTickers, pp.pair.GetPair(), pp.client)
	}
	return pp.bookTickerStream
}

func (pp *PairObserver) StartDepthsStream() *futures_streams.PartialDepthServeWithRate {
	if pp.depthsStream != nil {
		if pp.depths == nil {
			pp.depths = depth_types.New(degree, pp.pair.GetPair())
		}

		// Запускаємо потік для отримання оновлення depths
		pp.depthsStream = futures_streams.NewPartialDepthStreamWithRate(pp.pair.GetPair(), 5, 100, 1)
		pp.depthsStream.Start()
		futures_depths.Init(pp.depths, pp.client, pp.limit)
	}
	return pp.depthsStream
}

func (pp *PairObserver) StartRiskSignal() chan bool {
	if pp.riskEvent == nil {
		pp.riskEvent = make(chan bool, 1)
		go func() {
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-pp.riskEvent: // Чекаємо на спрацювання тригера
					riskPosition, err := pp.account.GetPositionRisk(pp.pair.GetPair())
					if err != nil {
						logrus.Errorf("Can't get position risk: %v, futures strategy", err)
						pp.stop <- os.Interrupt
						return
					}
					if len(riskPosition) != 1 {
						logrus.Errorf("Can't get correct position risk: %v, spot strategy", riskPosition)
						pp.stop <- os.Interrupt
						return
					}
					if (utils.ConvStrToFloat64(riskPosition[0].MarkPrice) -
						utils.ConvStrToFloat64(riskPosition[0].LiquidationPrice)/
							utils.ConvStrToFloat64(riskPosition[0].MarkPrice)) < 0.1 {
						logrus.Errorf("Risk position is too high: %v", riskPosition)
						pp.stop <- os.Interrupt
						return
					}
					pp.riskEvent <- true
				}
			}
		}()
	}
	return pp.riskEvent
}

func (pp *PairObserver) StartPriceByBookTickerSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
	if pp.bookTickers == nil {
		pp.bookTickers = book_ticker_types.New(degree)
	}
	bookTicker := pp.bookTickers.Get(pp.pair.GetPair())
	if bookTicker == nil {
		logrus.Errorf("Can't get bookTicker for %s when read for last price, spot strategy", pp.pair.GetPair())
		pp.stop <- os.Interrupt
		return
	}
	if pp.bookTickerStream == nil {
		pp.bookTickerStream = futures_streams.NewBookTickerStream(pp.pair.GetPair(), 1)
		pp.bookTickerStream.Start()
		futures_book_ticker.Init(pp.bookTickers, pp.pair.GetPair(), pp.client)
	}
	if pp.bookTickerEvent == nil {
		pp.bookTickerEvent = pp.StartBookTickersUpdateGuard()
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
				case <-pp.bookTickerEvent: // Чекаємо на спрацювання тригера на зміну ціни
					bookTicker := pp.bookTickers.Get(pp.pair.GetPair())
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

func (pp *PairObserver) StartPriceByDepthSignal() (
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
				case <-pp.depthEvent: // Чекаємо на спрацювання тригера на зміну ціни
					// Ціна купівлі
					ask,
						// Ціна продажу
						bid, err := pp.GetDepthAskBid()
					if err != nil {
						logrus.Errorf("Can't get depth for %s", pp.pair.GetPair())
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

// Запускаємо потік для оновлення ціни кожні updateTime
func (pp *PairObserver) StartPriceChangesSignal() (chan *pair_price_types.PairDelta, chan bool, chan bool) {
	if pp.priceChanges == nil && pp.priceUp == nil && pp.priceDown == nil {
		pp.priceChanges = make(chan *pair_price_types.PairDelta, 1)
		pp.priceUp = make(chan bool, 1)
		pp.priceDown = make(chan bool, 1)
		go func() {
			var (
				last_price float64
				price      *price_types.PriceChangeStats
			)
			price = price_types.New(degree)
			futures_price.Init(price, pp.client, pp.pair.GetPair())
			if priceVal := price.Get(&futures_price.SymbolPrice{Symbol: pp.pair.GetPair()}); priceVal != nil {
				last_price = utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price)
				logrus.Debugf("Start price for %s - %f", pp.pair.GetPair(), last_price)
			}
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-time.After(1 * time.Minute):
					price = price_types.New(degree)
					futures_price.Init(price, pp.client, pp.pair.GetPair())
					if priceVal := price.Get(&futures_price.SymbolPrice{Symbol: pp.pair.GetPair()}); priceVal != nil {
						if utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price) != 0 {
							current_price := utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price)
							logrus.Debugf("Current price for %s - %f", pp.pair.GetPair(), current_price)
							delta := (current_price - last_price) * 100 / last_price

							if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
								logrus.Debugf("Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
								pp.priceChanges <- &pair_price_types.PairDelta{
									Price:   utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price),
									Percent: utils.RoundToDecimalPlace(delta, 3)}
								if delta > 0 {
									pp.priceUp <- true
								} else {
									pp.priceDown <- true
								}
								last_price = current_price
							}
						}
					}
				}
			}
		}()
	}
	return pp.priceChanges, pp.priceUp, pp.priceDown
}

func (pp *PairObserver) StartBookTickersUpdateGuard() chan bool {
	if pp.bookTickerEvent == nil {
		pp.bookTickerEvent = futures_handlers.GetBookTickersUpdateGuard(pp.bookTickers, pp.bookTickerStream.GetDataChannel())
	}
	return pp.bookTickerEvent
}

func (pp *PairObserver) StartDepthsUpdateGuard() chan bool {
	if pp.depthEvent == nil {
		pp.depthEvent = futures_handlers.GetDepthsUpdateGuard(pp.depths, pp.depthsStream.GetDataChannel())
	}
	return pp.depthEvent
}

func (pp *PairObserver) StartWorkInPositionSignal(triggerEvent chan bool) (collectionOutEvent chan bool) { // Виходимо з накопичення
	if pp.pair.GetStage() != pair_types.InputIntoPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.InputIntoPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	if pp.collectionOutEvent == nil {
		pp.collectionOutEvent = make(chan bool, 1)

		go func() {
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-triggerEvent: // Чекаємо на спрацювання тригера
				case <-time.After(pp.pair.GetTakingPositionSleepingTime()): // Або просто чекаємо якийсь час
				}
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Ліміт на вхід в позицію, відсоток від балансу базової валюти
				LimitInputIntoPosition := pp.pair.GetLimitInputIntoPosition()
				// Ліміт на позицію, відсоток від балансу базової валюти
				LimitOnPosition := pp.pair.GetLimitOnPosition()
				// Верхня межа ціни купівлі
				boundAsk,
					// Нижня межа ціни продажу
					_, err := GetBound(pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість купівлі цільової валюти більша
				// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
				// - переходимо в режим спекуляції
				if targetBalance*boundAsk >= baseBalance*LimitInputIntoPosition ||
					targetBalance*boundAsk >= baseBalance*LimitOnPosition {
					pp.pair.SetStage(pair_types.WorkInPositionStage)
					collectionOutEvent <- true
					return
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.collectionOutEvent
}

func (pp *PairObserver) StopWorkInPositionSignal(triggerEvent chan bool) (positionOutEvent chan bool) { // Виходимо з спекуляції
	if pp.pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.WorkInPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	if positionOutEvent == nil {
		pp.positionOutEvent = make(chan bool, 1)

		go func() {
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-triggerEvent: // Чекаємо на спрацювання тригера
				case <-time.After(pp.pair.GetSleepingTime()): // Або просто чекаємо якийсь час
				}
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Ліміт на вхід в позицію, відсоток від балансу базової валюти
				LimitInputIntoPosition := pp.pair.GetLimitInputIntoPosition()
				// Ліміт на позицію, відсоток від балансу базової валюти
				LimitOnPosition := pp.pair.GetLimitOnPosition()
				// Верхня межа ціни купівлі
				_,
					// Нижня межа ціни продажу
					boundBid, err := GetBound(pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
				if targetBalance*boundBid >= baseBalance*LimitInputIntoPosition ||
					targetBalance*boundBid >= baseBalance*LimitOnPosition {
					pp.pair.SetStage(pair_types.OutputOfPositionStage)
					positionOutEvent <- true
					return
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.positionOutEvent
}

func NewPairObserver(
	client *futures.Client,
	account *futures_account.Account,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) *PairObserver {
	pp := &PairObserver{
		client:           client,
		pair:             pair,
		account:          account,
		stop:             stop,
		degree:           degree,
		limit:            limit,
		deltaUp:          deltaUp,
		deltaDown:        deltaDown,
		bookTickers:      nil,
		bookTickerStream: nil,
		depths:           nil,
		depthsStream:     nil,
		bookTickerEvent:  nil,
		depthEvent:       nil,
		priceChanges:     nil,
		priceUp:          nil,
		priceDown:        nil,
		askUp:            nil,
		askDown:          nil,
		bidUp:            nil,
		bidDown:          nil,
	}

	return pp
}
