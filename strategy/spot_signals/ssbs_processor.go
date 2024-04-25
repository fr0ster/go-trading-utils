package spot_signals

import (
	"fmt"
	_ "net/http/pprof"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_book_ticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairProcessor struct {
		client           *binance.Client
		pair             pairs_interfaces.Pairs
		account          *spot_account.Account
		stop             chan os.Signal
		bookTickers      *book_ticker_types.BookTickers
		bookTickerStream *spot_streams.BookTickerStream
		triggerEvent     chan bool
		buy              chan *pair_price_types.PairPrice
		sell             chan *pair_price_types.PairPrice
		up               chan *pair_price_types.PairPrice
		down             chan *pair_price_types.PairPrice
		wait             chan *pair_price_types.PairPrice
	}
)

func NewPairProcessor(
	client *binance.Client,
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	stop chan os.Signal,
	degree int) *PairProcessor {
	pp := &PairProcessor{
		client:           client,
		pair:             pair,
		account:          account,
		stop:             stop,
		bookTickers:      nil,
		bookTickerStream: spot_streams.NewBookTickerStream(pair.GetPair(), 1),
		triggerEvent:     make(chan bool),
		buy:              make(chan *pair_price_types.PairPrice, 1),
		sell:             make(chan *pair_price_types.PairPrice, 1),
		up:               make(chan *pair_price_types.PairPrice, 1),
		down:             make(chan *pair_price_types.PairPrice, 1),
	}
	pp.bookTickers = book_ticker_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	pp.bookTickerStream = spot_streams.NewBookTickerStream(pp.pair.GetPair(), 1)
	pp.bookTickerStream.Start()
	spot_book_ticker.Init(pp.bookTickers, pp.pair.GetPair(), client)
	pp.triggerEvent = spot_handlers.GetBookTickersUpdateGuard(pp.bookTickers, pp.bookTickerStream.DataChannel)
	return pp
}

func (pp *PairProcessor) GetBookTicker() *book_ticker_types.BookTicker {
	btk := pp.bookTickers.Get(pp.pair.GetPair())
	if btk == nil {
		return nil
	}
	return btk.(*book_ticker_types.BookTicker)
}

func (pp *PairProcessor) BuyOrSellSignal() (
	buy chan *pair_price_types.PairPrice,
	sell chan *pair_price_types.PairPrice) {
	return BuyOrSellSignal(pp.account, pp.bookTickers, pp.pair, pp.stop, pp.triggerEvent)
}

func (pp *PairProcessor) PriceSignal() (
	buy chan *pair_price_types.PairPrice,
	sell chan *pair_price_types.PairPrice,
	wait chan *pair_price_types.PairPrice) {
	return PriceSignal(pp.bookTickers, pp.pair, pp.stop, pp.triggerEvent)
}

func (pp *PairProcessor) Start() {
	// Запускаємо потік для отримання сигналів на купівлю та продаж
	pp.buy, pp.sell = BuyOrSellSignal(pp.account, pp.bookTickers, pp.pair, pp.stop, pp.triggerEvent)
	// Запускаємо потік для отримання сигналів росту та падіння ціни
	pp.up, pp.down, pp.wait = PriceSignal(pp.bookTickers, pp.pair, pp.stop, pp.triggerEvent)
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case <-pp.triggerEvent:
				fmt.Printf("%s, Signal, ask %v, bid %v\n", pp.pair.GetPair(), pp.GetBookTicker().AskPrice, pp.GetBookTicker().BidPrice)
			case <-pp.buy:
				fmt.Printf("%s, Signal buy\n", pp.pair.GetPair())
			case <-pp.sell:
				fmt.Printf("%s, Signal sell\n", pp.pair.GetPair())
			case <-pp.up:
				fmt.Printf("%s, Spot signal up\n", pp.pair.GetPair())
			case <-pp.down:
				fmt.Printf("%s, Spot signal down\n", pp.pair.GetPair())
			}
		}
	}()
}

func (pp *PairProcessor) StartBuyOrSellSignal() {
	// Запускаємо потік для отримання сигналів на купівлю та продаж
	pp.buy, pp.sell = BuyOrSellSignal(pp.account, pp.bookTickers, pp.pair, pp.stop, pp.triggerEvent)
}

func (pp *PairProcessor) StartBuyOrSellHandler() {
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case <-pp.triggerEvent:
				logrus.Debugf("%s, Signal, ask %v, bid %v\n", pp.pair.GetPair(), pp.GetBookTicker().AskPrice, pp.GetBookTicker().BidPrice)
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
	pp.buy, pp.sell, pp.wait = PriceSignal(pp.bookTickers, pp.pair, pp.stop, pp.triggerEvent)
}

func (pp *PairProcessor) StartPriceHandler() {
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case <-pp.triggerEvent:
				logrus.Debugf("%s, Signal, ask %v, bid %v\n", pp.pair.GetPair(), pp.GetBookTicker().AskPrice, pp.GetBookTicker().BidPrice)
			case price := <-pp.up:
				fmt.Printf("%s, Signal up, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			case price := <-pp.down:
				fmt.Printf("%s, Signal down, price %f, quantity %f \n", pp.pair.GetPair(), price.Price, price.Quantity)
			}
		}
	}()
}
