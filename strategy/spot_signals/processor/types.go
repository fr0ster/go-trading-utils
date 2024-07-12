package processor

import (
	"time"

	"github.com/adshao/go-binance/v2"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
)

type (
	PairProcessor struct {
		client       *binance.Client
		exchangeInfo *exchange_types.ExchangeInfo
		symbol       *binance.Symbol
		baseSymbol   string
		targetSymbol string
		notional     float64
		StepSize     float64
		maxQty       float64
		minQty       float64
		tickSize     float64
		maxPrice     float64
		minPrice     float64

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

		progression             pairs_types.ProgressionType
		GetDelta                progressions.DeltaType
		NthTerm                 progressions.NthTermType
		Sum                     progressions.SumType
		FindNthTerm             progressions.FindNthTermType
		FindLengthOfProgression progressions.FindLengthOfProgressionType
		FindProgressionTthTerm  progressions.FindCubicProgressionTthTermType

		depth *depth_types.Depth
	}
)

const (
	errorMsg = "Error: %v"
)
