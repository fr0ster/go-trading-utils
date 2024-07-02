package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
	"github.com/google/btree"
)

const (
	errorMsg = "Error: %v"
)

type (
	PairProcessor struct {
		client        *futures.Client
		exchangeInfo  *exchange_types.ExchangeInfo
		symbol        *futures.Symbol
		baseSymbol    string
		targetSymbol  string
		notional      float64
		stepSizeDelta float64
		minSteps      int
		up            *btree.BTree
		down          *btree.BTree

		stop chan struct{}

		pairInfo           *symbol_types.FuturesSymbol
		orderTypes         map[futures.OrderType]bool
		degree             int
		sleepingTime       time.Duration
		timeOut            time.Duration
		limitOnPosition    float64
		limitOnTransaction float64
		UpBoundPercent     float64
		UpBound            float64
		LowBoundPercent    float64
		LowBound           float64
		leverage           int
		marginType         pairs_types.MarginType
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
	}
)
