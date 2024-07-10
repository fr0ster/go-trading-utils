package processor

import (
	"time"

	"github.com/adshao/go-binance/v2"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	nextPriceFunc    func(float64, int) float64
	nextQuantityFunc func(float64, int) float64
	testFunc         func(float64, float64) bool
	Functions        struct {
		NextPriceUp      nextPriceFunc
		NextPriceDown    nextPriceFunc
		NextQuantityUp   nextQuantityFunc
		NextQuantityDown nextQuantityFunc
		TestUp           testFunc
		TestDown         testFunc
	}
	PairProcessor struct {
		client        *binance.Client
		exchangeInfo  *exchange_types.ExchangeInfo
		symbol        *binance.Symbol
		baseSymbol    string
		targetSymbol  string
		notional      float64
		stepSizeDelta float64

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

		stop chan struct{}

		pairInfo           *symbol_types.SpotSymbol
		orderTypes         map[string]bool
		degree             int
		sleepingTime       time.Duration
		timeOut            time.Duration
		limitOnPosition    float64
		limitOnTransaction float64
		UpBound            float64
		LowBound           float64
		callbackRate       float64

		deltaPrice    float64
		deltaQuantity float64

		testUp           testFunc
		testDown         testFunc
		NextPriceUp      nextPriceFunc
		NextPriceDown    nextPriceFunc
		NextQuantityUp   nextQuantityFunc
		NextQuantityDown nextQuantityFunc
	}
)

const (
	errorMsg = "Error: %v"
)
