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
		buy              chan *pair_price_types.PairPrice
		sell             chan *pair_price_types.PairPrice
		up               chan *pair_price_types.PairPrice
		down             chan *pair_price_types.PairPrice
		wait             chan *pair_price_types.PairPrice
		stop             chan os.Signal
	}
)

func (pp *PairProcessor) GetBookTicker() *book_ticker_types.BookTicker {
	btk := pp.bookTickers.Get(pp.pair.GetPair())
	if btk == nil {
		return nil
	}
	return btk.(*book_ticker_types.BookTicker)
}

func (pp *PairProcessor) GetDepth() *depth_types.Depth {
	return pp.depths
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

func (pp *PairProcessor) BuyOrSellByBookTickerSignal() (
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

func (pp *PairProcessor) BuyOrSellByDepthSignal() (
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

func (pp *PairProcessor) PriceSignal() (
	buy chan *pair_price_types.PairPrice,
	sell chan *pair_price_types.PairPrice,
	wait chan *pair_price_types.PairPrice) {
	return PriceSignal(pp.bookTickers, pp.pair, pp.stop, pp.bookTickerEvent)
}

func (pp *PairProcessor) Start() {
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case <-pp.bookTickerEvent:
				fmt.Printf("%s, Signal bookTickerEvent, ask %f, bid %f\n", pp.pair.GetPair(), pp.GetBookTicker().AskPrice, pp.GetBookTicker().BidPrice)
			case <-pp.depthEvent:
				minAsk := pp.depths.GetAsks().Min().(*pair_price_types.PairPrice)
				maxBid := pp.depths.GetBids().Max().(*pair_price_types.PairPrice)
				fmt.Printf("%s, Signal depthEvent, ask %f, bid %f\n", pp.pair.GetPair(), minAsk.Price, maxBid.Price)
			}
		}
	}()
}

func (pp *PairProcessor) StartBookTickersUpdateGuard() {
	pp.bookTickerEvent = spot_handlers.GetBookTickersUpdateGuard(pp.bookTickers, pp.bookTickerStream.GetDataChannel())
}

func (pp *PairProcessor) StartDepthsUpdateGuard() {
	pp.depthEvent = spot_handlers.GetDepthsUpdateGuard(pp.depths, pp.depthsStream.GetDataChannel())
}

func (pp *PairProcessor) StartBuyOrSellByBookTickerSignal() {
	// Запускаємо потік для отримання сигналів на купівлю та продаж
	pp.buy, pp.sell = pp.BuyOrSellByBookTickerSignal()
}

func (pp *PairProcessor) StartBuyOrSellByDepthSignal() {
	// Запускаємо потік для отримання сигналів на купівлю та продаж
	pp.buy, pp.sell = pp.BuyOrSellByDepthSignal()
}

func (pp *PairProcessor) StartBuyOrSellHandler() {
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case price := <-pp.buy:
				fmt.Printf("%s, Signal buy, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			case price := <-pp.sell:
				fmt.Printf("%s, Signal sell, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			}
		}
	}()
}

func (pp *PairProcessor) StartPriceSignal() {
	// Запускаємо потік для отримання сигналів на купівлю та продаж
	pp.up, pp.up, pp.wait = PriceSignal(pp.bookTickers, pp.pair, pp.stop, pp.bookTickerEvent)
}

func (pp *PairProcessor) StartPriceHandler() {
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case price := <-pp.up:
				fmt.Printf("%s, Signal up, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			case price := <-pp.down:
				fmt.Printf("%s, Signal down, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			case price := <-pp.wait:
				fmt.Printf("%s, Signal wait, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			}
		}
	}()
}

func (pp *PairProcessor) StartPriceSignal1() {
	var lastPrice float64
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case <-pp.bookTickerStream.GetEventChannel(): // Чекаємо на спрацювання тригера
				// Ціна купівлі
				ask := pp.bookTickers.Get(pp.pair.GetPair()).(*book_ticker_types.BookTicker).AskPrice
				// Ціна продажу
				bid := pp.bookTickers.Get(pp.pair.GetPair()).(*book_ticker_types.BookTicker).BidPrice
				if lastPrice == 0 {
					lastPrice = (ask + bid) / 2
				}
				currentPrice := (ask + bid) / 2
				if currentPrice > lastPrice {
					pp.up <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
					lastPrice = currentPrice
				} else if currentPrice < lastPrice {
					pp.down <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
					lastPrice = currentPrice
				} else {
					pp.wait <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
				}
			}
			time.Sleep(pp.pair.GetSleepingTime())
		}
	}()
}

func NewPairProcessor(
	client *binance.Client,
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	stop chan os.Signal) *PairProcessor {
	pp := &PairProcessor{
		client:           client,
		pair:             pair,
		account:          account,
		stop:             stop,
		degree:           degree,
		limit:            limit,
		bookTickers:      nil,
		bookTickerStream: spot_streams.NewBookTickerStream(pair.GetPair(), 1),
		depths:           nil,
		depthsStream:     spot_streams.NewDepthStream(pair.GetPair(), true, 1),
		bookTickerEvent:  make(chan bool),
		depthEvent:       make(chan bool),
		buy:              make(chan *pair_price_types.PairPrice, 1),
		sell:             make(chan *pair_price_types.PairPrice, 1),
		up:               make(chan *pair_price_types.PairPrice, 1),
		down:             make(chan *pair_price_types.PairPrice, 1),
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
