package spot_signals

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_book_ticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairBookTickersObserver struct {
		client          *binance.Client
		degree          int
		limit           int
		account         *spot_account.Account
		exchangeInfo    *exchange_info.ExchangeInfo
		data            *book_ticker_types.BookTickers
		bookTickerEvent chan *binance.WsBookTickerEvent
		event           chan bool
		stop            chan struct{}
		deltaUp         float64
		deltaDown       float64
		askUp           chan *pair_price_types.AskBid
		askDown         chan *pair_price_types.AskBid
		bidUp           chan *pair_price_types.AskBid
		bidDown         chan *pair_price_types.AskBid
		sleepingTime    time.Duration
		timeOut         time.Duration
		symbol          *binance.Symbol
	}
)

func (pp *PairBookTickersObserver) GetBookTickers() *book_ticker_types.BookTicker {
	btk := pp.data.Get(pp.symbol.Symbol)
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

		ticker := time.NewTicker(pp.timeOut)
		lastResponse := time.Now()
		// Запускаємо потік для отримання оновлення bookTickers
		logrus.Debugf("Spot, Start stream for %v Klines", pp.symbol.Symbol)
		pp.bookTickerEvent = make(chan *binance.WsBookTickerEvent, 1)
		wsHandler := func(event *binance.WsBookTickerEvent) {
			lastResponse = time.Now()
			pp.bookTickerEvent <- event
		}
		resetEvent := make(chan bool, 1)
		wsErrorHandler := func(err error) {
			resetEvent <- true
		}
		var stopC chan struct{}
		_, stopC, _ = binance.WsBookTickerServe(pp.symbol.Symbol, wsHandler, wsErrorHandler)
		go func() {
			for {
				select {
				case <-resetEvent:
					stopC <- struct{}{}
					_, stopC, _ = binance.WsBookTickerServe(pp.symbol.Symbol, wsHandler, wsErrorHandler)
				case <-ticker.C:
					if time.Since(lastResponse) > pp.timeOut {
						stopC <- struct{}{}
						_, stopC, _ = binance.WsBookTickerServe(pp.symbol.Symbol, wsHandler, wsErrorHandler)
					}
				}
			}
		}()
		spot_book_ticker.Init(pp.data, pp.symbol.Symbol, pp.client)
	}
	return pp.bookTickerEvent
}

func (pp *PairBookTickersObserver) GetAskBid() (bid float64, ask float64, err error) {
	btk := pp.data.Get(pp.symbol.Symbol)
	if btk == nil {
		err = fmt.Errorf("can't get bookTicker for %s", pp.symbol.Symbol)
		return
	}
	ask = btk.(*book_ticker_types.BookTicker).AskPrice
	bid = btk.(*book_ticker_types.BookTicker).BidPrice
	return
}

func (pp *PairBookTickersObserver) StartPriceChangesSignal() (
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid) {
	bookTicker := pp.data.Get(pp.symbol.Symbol)
	if pp.data == nil {
		pp.data = book_ticker_types.New(degree)
	}
	if bookTicker == nil {
		logrus.Errorf("Can't get bookTicker for %s when read for last price, spot strategy", pp.symbol.Symbol)
		close(pp.stop)
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
					close(pp.stop)
					return
				case <-pp.event: // Чекаємо на спрацювання тригера на зміну ціни
					bookTicker := pp.data.Get(pp.symbol.Symbol)
					if bookTicker == nil {
						logrus.Errorf("Can't get bookTicker for %s", pp.symbol.Symbol)
						close(pp.stop)
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
					logrus.Debugf("Spot, Ask is %f, Last Ask is %f, Delta Ask is%f%%, Bid is %f, Last Bid is %f, Delta Bid is %f%%",
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

func (pp *PairBookTickersObserver) StartUpdateGuard() chan bool {
	if pp.event == nil {
		if pp.bookTickerEvent == nil {
			pp.StartStream()
		}
		pp.event = spot_handlers.GetBookTickersUpdateGuard(pp.data, pp.bookTickerEvent)
	}
	return pp.event
}

func (pp *PairBookTickersObserver) GetMinQuantity(price float64) float64 {
	return utils.ConvStrToFloat64(pp.symbol.NotionalFilter().MinNotional) / price
}

func (pp *PairBookTickersObserver) GetMaxQuantity(price float64) float64 {
	return utils.ConvStrToFloat64(pp.symbol.NotionalFilter().MaxNotional) / price
}

func (pp *PairBookTickersObserver) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairBookTickersObserver) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func NewPairBookTickersObserver(
	client *binance.Client,
	symbol string,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan struct{}) (pp *PairBookTickersObserver, err error) {
	pp = &PairBookTickersObserver{
		client:          client,
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
		sleepingTime:    1 * time.Second,
		timeOut:         1 * time.Hour,
	}
	pp.exchangeInfo = exchange_info.New()
	err = spot_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}
	if symbol := pp.exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: symbol}); symbol != nil {
		pp.symbol, err = symbol.(*symbol_info.SpotSymbol).GetSpotSymbol()
		if err != nil {
			logrus.Errorf(errorMsg, err)
			return
		}
	}
	pp.account, err = spot_account.New(pp.client, []string{pp.symbol.QuoteAsset, pp.symbol.BaseAsset})
	if err != nil {
		return
	}

	return
}
