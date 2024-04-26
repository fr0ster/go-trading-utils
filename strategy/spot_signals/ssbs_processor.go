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
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairProcessor struct {
		client           *binance.Client
		pair             pairs_interfaces.Pairs
		account          *spot_account.Account
		bookTickers      *book_ticker_types.BookTickers
		bookTickerStream *spot_streams.BookTickerStream
		degree           int
		limit            int
		depths           *depth_types.Depth
		depthsStream     *spot_streams.DepthStream
		bookTickerEvent  chan bool
		depthEvent       chan bool
		stop             chan os.Signal
		deltaUp          float64
		deltaDown        float64
	}
)

func (pp *PairProcessor) GetBookTicker() *book_ticker_types.BookTicker {
	btk := pp.bookTickers.Get(pp.pair.GetPair())
	if btk == nil {
		return nil
	}
	return btk.(*book_ticker_types.BookTicker)
}

func (pp *PairProcessor) GetBookTickerStream() *spot_streams.BookTickerStream {
	return pp.bookTickerStream
}

func (pp *PairProcessor) GetDepth() *depth_types.Depth {
	return pp.depths
}

func (pp *PairProcessor) GetDepthStream() *spot_streams.DepthStream {
	return pp.depthsStream
}

func (pp *PairProcessor) GetBookTickerAskBid() (bid float64, ask float64, err error) {
	btk := pp.bookTickers.Get(pp.pair.GetPair())
	if btk == nil {
		err = fmt.Errorf("can't get bookTicker for %s", pp.pair.GetPair())
		return
	}
	ask = btk.(*book_ticker_types.BookTicker).AskPrice
	bid = btk.(*book_ticker_types.BookTicker).BidPrice
	return
}

func (pp *PairProcessor) GetDepthAskBid() (bid float64, ask float64, err error) {
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

func (pp *PairProcessor) StartBuyOrSellByBookTickerSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	buyEvent = make(chan *pair_price_types.PairPrice, 1)
	sellEvent = make(chan *pair_price_types.PairPrice, 1)
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
	return
}

func (pp *PairProcessor) StartBuyOrSellByDepthSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	buyEvent = make(chan *pair_price_types.PairPrice, 1)
	sellEvent = make(chan *pair_price_types.PairPrice, 1)
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
	return
}

func (pp *PairProcessor) StartPriceSignal() (
	up chan *pair_price_types.PairPrice,
	down chan *pair_price_types.PairPrice,
	wait chan *pair_price_types.PairPrice) {
	up = make(chan *pair_price_types.PairPrice, 1)
	down = make(chan *pair_price_types.PairPrice, 1)
	wait = make(chan *pair_price_types.PairPrice, 1)
	bookTicker := pp.bookTickers.Get(pp.pair.GetPair())
	if bookTicker == nil {
		logrus.Errorf("Can't get bookTicker for %s when read for last price, spot strategy", pp.pair.GetPair())
		pp.stop <- os.Interrupt
		return
	}
	go func() {
		var last_bid, last_ask, lastPrice float64
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
				bid := bookTicker.(*book_ticker_types.BookTicker).AskPrice
				if last_bid == 0 || last_ask == 0 {
					last_bid = bid
					last_ask = ask
				}
				if lastPrice == 0 {
					lastPrice = (ask + bid) / 2
				}
				if ask == last_ask && bid == last_bid {
					wait <- &pair_price_types.PairPrice{
						Price: (ask + bid) / 2,
					}
				} else {
					currentPrice := (ask + bid) / 2
					if currentPrice > lastPrice*(1+pp.deltaUp) {
						up <- &pair_price_types.PairPrice{
							Price: currentPrice,
						}
						lastPrice = currentPrice
					} else if currentPrice < lastPrice*(1-pp.deltaDown) {
						down <- &pair_price_types.PairPrice{
							Price: currentPrice,
						}
						lastPrice = currentPrice
					} else {
						wait <- &pair_price_types.PairPrice{
							Price: currentPrice,
						}
					}
				}
			}
			time.Sleep(pp.pair.GetSleepingTime())
		}
	}()
	return
}

func (pp *PairProcessor) StartBookTickersUpdateGuard() chan bool {
	pp.bookTickerEvent = spot_handlers.GetBookTickersUpdateGuard(pp.bookTickers, pp.bookTickerStream.GetDataChannel())
	return pp.bookTickerEvent
}

func (pp *PairProcessor) StartDepthsUpdateGuard() chan bool {
	pp.depthEvent = spot_handlers.GetDepthsUpdateGuard(pp.depths, pp.depthsStream.GetDataChannel())
	return pp.depthEvent
}

func NewPairProcessor(
	client *binance.Client,
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) *PairProcessor {
	pp := &PairProcessor{
		client:           client,
		pair:             pair,
		account:          account,
		stop:             stop,
		degree:           degree,
		limit:            limit,
		deltaUp:          deltaUp,
		deltaDown:        deltaDown,
		bookTickers:      nil,
		bookTickerStream: spot_streams.NewBookTickerStream(pair.GetPair(), 1),
		depths:           nil,
		depthsStream:     spot_streams.NewDepthStream(pair.GetPair(), true, 1),
		bookTickerEvent:  make(chan bool),
		depthEvent:       make(chan bool),
	}
	pp.bookTickers = book_ticker_types.New(degree)
	pp.depths = depth_types.New(degree, pp.pair.GetPair())

	// Запускаємо потік для отримання оновлення bookTickers
	pp.bookTickerStream = spot_streams.NewBookTickerStream(pp.pair.GetPair(), 1)
	pp.bookTickerStream.Start()
	spot_book_ticker.Init(pp.bookTickers, pp.pair.GetPair(), client)

	// Запускаємо потік для отримання оновлення depths
	pp.depthsStream = spot_streams.NewDepthStream(pp.pair.GetPair(), true, 1)
	pp.depthsStream.Start()
	spot_depths.Init(pp.depths, client, pp.limit)

	return pp
}
