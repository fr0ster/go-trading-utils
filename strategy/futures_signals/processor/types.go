package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/google/btree"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/fr0ster/go-trading-utils/types/depth/types"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
)

const (
	errorMsg    = "Error: %v"
	repeatTimes = 3
)

type (
	PairProcessor struct {
		// Дані про клієнта
		client *futures.Client
		// Налаштування та обмеження, реалізація
		orderTypes map[futures.OrderType]bool
		degree     int
		timeOut    time.Duration

		// Дані про біржу
		exchangeInfo *exchange_types.ExchangeInfo

		// Дані про пару
		symbol       *futures.Symbol
		pairInfo     *symbol_types.FuturesSymbol
		baseSymbol   string
		targetSymbol string

		// Дані про обмеження на пару
		notional float64
		StepSize float64
		maxQty   float64
		minQty   float64
		tickSize float64
		maxPrice float64
		minPrice float64

		// канал зупинки
		stop chan struct{}

		limitOnPosition    types.PriceType
		limitOnTransaction float64

		// Дінаміка ціни, використовувалось тіко для grid_v3
		minSteps        int
		up              *btree.BTree
		down            *btree.BTree
		UpBoundPercent  float64
		UpBound         types.PriceType
		LowBoundPercent float64
		LowBound        types.PriceType
		deltaPrice      types.PriceType
		deltaQuantity   types.QuantityType

		// Прогресії, використовувалось тіко для grid_v3
		progression             pairs_types.ProgressionType
		GetDelta                progressions.DeltaType
		NthTerm                 progressions.NthTermType
		Sum                     progressions.SumType
		FindNthTerm             progressions.FindNthTermType
		FindLengthOfProgression progressions.FindLengthOfProgressionType
		FindProgressionTthTerm  progressions.FindCubicProgressionTthTermType

		// Дані про позицію
		leverage     int
		marginType   pairs_types.MarginType
		callbackRate float64

		// Дані про стакан
		depth *depth_types.Depth
	}
)
