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
	spot_price "github.com/fr0ster/go-trading-utils/binance/spot/markets/price"
	"github.com/fr0ster/go-trading-utils/utils"

	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
	price_types "github.com/fr0ster/go-trading-utils/types/price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairObserver struct {
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
		priceChanges     chan *pair_price_types.PairDelta
		stop             chan os.Signal
		deltaUp          float64
		deltaDown        float64
		buyEvent         chan *pair_price_types.PairPrice
		sellEvent        chan *pair_price_types.PairPrice
		askUp            chan *pair_price_types.AskBid
		askDown          chan *pair_price_types.AskBid
		bidUp            chan *pair_price_types.AskBid
		bidDown          chan *pair_price_types.AskBid
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

func (pp *PairObserver) GetDepth() *depth_types.Depth {
	return pp.depths
}

func (pp *PairObserver) GetDepthStream() *spot_streams.DepthStream {
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

func (pp *PairObserver) StartBuyOrSellByBookTickerSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
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
	return
}

func (pp *PairObserver) StartBuyOrSellByDepthSignal() (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
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
	return
}

func (pp *PairObserver) StartPriceByBookTickerSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
	bookTicker := pp.bookTickers.Get(pp.pair.GetPair())
	if bookTicker == nil {
		logrus.Errorf("Can't get bookTicker for %s when read for last price, spot strategy", pp.pair.GetPair())
		pp.stop <- os.Interrupt
		return
	}
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
	return pp.askUp, pp.askDown, pp.bidUp, pp.bidDown
}

func (pp *PairObserver) StartPriceByDepthSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
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
	return pp.askUp, pp.askDown, pp.bidUp, pp.bidDown
}

// Запускаємо потік для оновлення ціни кожні updateTime
func (pp *PairObserver) StartPriceChangesSignal() chan *pair_price_types.PairDelta {

	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case <-time.After(1 * time.Minute):
				price := price_types.New(degree)
				spot_price.Init(price, pp.client, pp.pair.GetPair())
				if priceVal := price.Get(&spot_price.SymbolTicker{Symbol: pp.pair.GetPair()}); priceVal != nil {
					if utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).PriceChange) != 0 {
						pp.priceChanges <- &pair_price_types.PairDelta{
							Price:   utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice),
							Percent: utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).PriceChangePercent)}
					}
				}
			}
		}
	}()
	return pp.priceChanges
}

func (pp *PairObserver) StartBookTickersUpdateGuard() chan bool {
	pp.bookTickerEvent = spot_handlers.GetBookTickersUpdateGuard(pp.bookTickers, pp.bookTickerStream.GetDataChannel())
	return pp.bookTickerEvent
}

func (pp *PairObserver) StartDepthsUpdateGuard() chan bool {
	pp.depthEvent = spot_handlers.GetDepthsUpdateGuard(pp.depths, pp.depthsStream.GetDataChannel())
	return pp.depthEvent
}

func (pp *PairObserver) StartWorkInPositionSignal(triggerEvent chan bool) (
	collectionOutEvent chan bool) { // Виходимо з накопичення
	if pp.pair.GetStage() != pair_types.InputIntoPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.InputIntoPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	collectionOutEvent = make(chan bool, 1)

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
	return
}

func (pp *PairObserver) StopWorkInPositionSignal(triggerEvent chan bool) (
	workingOutEvent chan bool) { // Виходимо з накопичення
	if pp.pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.WorkInPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	workingOutEvent = make(chan bool, 1)

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
	return
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
		stop:             stop,
		degree:           degree,
		limit:            limit,
		deltaUp:          deltaUp,
		deltaDown:        deltaDown,
		bookTickers:      nil,
		bookTickerStream: spot_streams.NewBookTickerStream(pair.GetPair(), 1),
		depths:           nil,
		depthsStream:     spot_streams.NewDepthStream(pair.GetPair(), true, 1),
		bookTickerEvent:  make(chan bool, 1),
		depthEvent:       make(chan bool, 1),
		priceChanges:     make(chan *pair_price_types.PairDelta, 1),
		askUp:            make(chan *pair_price_types.AskBid, 1),
		askDown:          make(chan *pair_price_types.AskBid, 1),
		bidUp:            make(chan *pair_price_types.AskBid, 1),
		bidDown:          make(chan *pair_price_types.AskBid, 1),
	}
	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
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

	return
}
