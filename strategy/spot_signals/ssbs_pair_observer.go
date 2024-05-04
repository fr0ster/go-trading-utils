package spot_signals

import (
	"fmt"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_book_ticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	spot_depths "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	spot_kline "github.com/fr0ster/go-trading-utils/binance/spot/markets/kline"
	spot_price "github.com/fr0ster/go-trading-utils/binance/spot/markets/price"

	"github.com/fr0ster/go-trading-utils/utils"

	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
	price_types "github.com/fr0ster/go-trading-utils/types/price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairObserver struct {
		client             *binance.Client
		pair               pairs_interfaces.Pairs
		degree             int
		limit              int
		account            *spot_account.Account
		bookTickers        *book_ticker_types.BookTickers
		bookTickerStream   *spot_streams.BookTickerStream
		bookTickerEvent    chan bool
		depths             *depth_types.Depth
		depthsStream       *spot_streams.DepthStream
		depthEvent         chan bool
		klines             *kline_types.Klines
		klineStream        *spot_streams.KlineStream
		klineEvent         chan bool
		priceChanges       chan *pair_price_types.PairDelta
		collectionOutEvent chan bool
		workingOutEvent    chan bool
		priceUp            chan bool
		priceDown          chan bool
		stop               chan os.Signal
		deltaUp            float64
		deltaDown          float64
		buyEvent           chan *pair_price_types.PairPrice
		sellEvent          chan *pair_price_types.PairPrice
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

func (pp *PairObserver) GetBookTickerStream() *spot_streams.BookTickerStream {
	return pp.bookTickerStream
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

func (pp *PairObserver) StartBookTickerStream() *spot_streams.BookTickerStream {
	if pp.bookTickerStream != nil {
		if pp.bookTickers == nil {
			pp.bookTickers = book_ticker_types.New(degree)
		}

		// Запускаємо потік для отримання оновлення bookTickers
		pp.bookTickerStream = spot_streams.NewBookTickerStream(pp.pair.GetPair(), 1)
		pp.bookTickerStream.Start()
		spot_book_ticker.Init(pp.bookTickers, pp.pair.GetPair(), pp.client)
	}
	return pp.bookTickerStream
}

func (pp *PairObserver) StartBuyOrSellByBookTickerSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	if pp.bookTickerStream == nil {
		pp.StartBookTickerStream()
	}
	if pp.bookTickerEvent == nil {
		pp.StartBookTickersUpdateGuard()
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
				case <-pp.bookTickerEvent: // Чекаємо на спрацювання тригера на зміну bookTicker
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
					bookTicker := pp.bookTickers.Get(pp.pair.GetPair())
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

func (pp *PairObserver) StartPriceByBookTickerSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
	bookTicker := pp.bookTickers.Get(pp.pair.GetPair())
	if pp.bookTickers == nil {
		pp.bookTickers = book_ticker_types.New(degree)
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

func (pp *PairObserver) StartBookTickersUpdateGuard() chan bool {
	if pp.bookTickerEvent == nil {
		if pp.bookTickerStream == nil {
			pp.StartBookTickerStream()
		}
		pp.bookTickerEvent = spot_handlers.GetBookTickersUpdateGuard(pp.bookTickers, pp.bookTickerStream.GetDataChannel())
	}
	return pp.bookTickerEvent
}

func (pp *PairObserver) GetDepth() *depth_types.Depth {
	return pp.depths
}

func (pp *PairObserver) GetDepthStream() *spot_streams.DepthStream {
	return pp.depthsStream
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

func (pp *PairObserver) StartDepthsStream() *spot_streams.DepthStream {
	if pp.depthsStream != nil {
		if pp.depths == nil {
			pp.depths = depth_types.New(degree, pp.pair.GetPair())
		}

		// Запускаємо потік для отримання оновлення depths
		pp.depthsStream = spot_streams.NewDepthStream(pp.pair.GetPair(), true, 1)
		pp.depthsStream.Start()
		spot_depths.Init(pp.depths, pp.client, pp.limit)
	}
	return pp.depthsStream
}

func (pp *PairObserver) StartBuyOrSellByDepthSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	if pp.depthsStream == nil {
		pp.StartDepthsStream()
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
				case <-pp.depthEvent: // Чекаємо на спрацювання тригера на зміну bookTicker
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
					minAsk := pp.depths.GetAsks().Min()
					maxBid := pp.depths.GetBids().Max()
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

func (pp *PairObserver) StartDepthsUpdateGuard() chan bool {
	if pp.depthEvent == nil {
		if pp.depthsStream == nil {
			pp.StartDepthsStream()
		}
		pp.depthEvent = spot_handlers.GetDepthsUpdateGuard(pp.depths, pp.depthsStream.GetDataChannel())
	}
	return pp.depthEvent
}

func (pp *PairObserver) GetKlines() *kline_types.Klines {
	return pp.klines
}

func (pp *PairObserver) GetKlineStream() *spot_streams.KlineStream {
	return pp.klineStream
}

func (pp *PairObserver) StartKlinesStream() *spot_streams.KlineStream {
	if pp.klineStream != nil {
		if pp.klines == nil {
			pp.klines = kline_types.New(degree)
		}

		// Запускаємо потік для отримання оновлення depths
		pp.klineStream = spot_streams.NewKlineStream(pp.pair.GetPair(), "1m", 1)
		pp.klineStream.Start()
		spot_kline.Init(pp.klines, pp.client, pp.pair.GetPair())
	}
	return pp.klineStream
}

// Запускаємо потік для оновлення ціни кожні updateTime
func (pp *PairObserver) StartPriceChangesSignal() (chan *pair_price_types.PairDelta, chan bool, chan bool) {
	if pp.priceChanges == nil && pp.priceUp == nil && pp.priceDown == nil {
		pp.priceChanges = make(chan *pair_price_types.PairDelta, 1)
		pp.priceUp = make(chan bool, 1)
		pp.priceDown = make(chan bool, 1)
		go func() {
			var (
				price      *price_types.PriceChangeStats
				last_price float64
			)
			price = price_types.New(degree)
			spot_price.Init(price, pp.client, pp.pair.GetPair())
			if priceVal := price.Get(&spot_price.SymbolTicker{Symbol: pp.pair.GetPair()}); priceVal != nil {
				last_price = utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice)
				logrus.Debugf("Start price for %s - %f", pp.pair.GetPair(), last_price)
			}
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-time.After(1 * time.Minute):
					price = price_types.New(degree)
					spot_price.Init(price, pp.client, pp.pair.GetPair())
					if priceVal := price.Get(&spot_price.SymbolTicker{Symbol: pp.pair.GetPair()}); priceVal != nil {
						if utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice) != 0 {
							current_price := utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice)
							delta := (current_price - last_price) * 100 / last_price
							logrus.Debugf("Current price for %s - %f, delta - %f", pp.pair.GetPair(), current_price, delta)
							if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
								logrus.Debugf("Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
								pp.priceChanges <- &pair_price_types.PairDelta{
									Price:   utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice),
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

func (pp *PairObserver) StartWorkInPositionSignal(triggerEvent chan bool) (
	collectionOutEvent chan bool) { // Виходимо з накопичення
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
				pp.pair.SetCurrentBalance(baseBalance)
				pp.pair.SetCurrentPositionBalance(baseBalance * pp.pair.GetLimitOnPosition())
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Верхня межа ціни купівлі
				boundAsk, err := GetAskBound(pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість купівлі цільової валюти більша
				// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
				// - переходимо в режим спекуляції
				if targetBalance*boundAsk >= baseBalance*pp.pair.GetLimitInputIntoPosition() {
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

func (pp *PairObserver) StopWorkInPositionSignal(triggerEvent chan bool) (
	workingOutEvent chan bool) { // Виходимо з спекуляції
	if pp.pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.WorkInPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	if pp.workingOutEvent == nil {
		pp.workingOutEvent = make(chan bool, 1)

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
				pp.pair.SetCurrentBalance(baseBalance)
				pp.pair.SetCurrentPositionBalance(baseBalance * pp.pair.GetLimitOnPosition())
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Нижня межа ціни продажу
				boundBid, err := GetBidBound(pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
				if targetBalance*boundBid >= baseBalance*pp.pair.GetLimitOutputOfPosition() {
					pp.pair.SetStage(pair_types.OutputOfPositionStage)
					workingOutEvent <- true
					return
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.workingOutEvent
}

func NewPairObserver(
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) (pp *PairObserver, err error) {
	pp = &PairObserver{
		client:           client,
		pair:             pair,
		account:          nil,
		bookTickers:      nil,
		bookTickerStream: nil,
		bookTickerEvent:  nil,
		depths:           nil,
		depthsStream:     nil,
		depthEvent:       nil,
		klines:           nil,
		klineStream:      nil,
		klineEvent:       nil,
		stop:             stop,
		degree:           degree,
		limit:            limit,
		deltaUp:          deltaUp,
		deltaDown:        deltaDown,
		priceChanges:     nil,
		priceUp:          nil,
		priceDown:        nil,
		askUp:            nil,
		askDown:          nil,
		bidUp:            nil,
		bidDown:          nil,
	}
	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
