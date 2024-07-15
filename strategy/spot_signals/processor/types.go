package processor

import (
	"time"

	"github.com/adshao/go-binance/v2"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
)

type (
	PairProcessor struct {
		// Дані про клієнта
		client *binance.Client
		// Налаштування та обмеження, реалізація
		orderTypes map[binance.OrderType]bool
		degree     int
		timeOut    time.Duration

		// Дані про біржу
		exchangeInfo *exchange_types.ExchangeInfo

		// Дані про пару
		symbol       *binance.Symbol
		pairInfo     *symbol_types.SpotSymbol
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

		limitOnPosition    float64
		limitOnTransaction float64

		// Дінаміка ціни, використовувалось тіко для grid_v3
		UpBoundPercent  float64
		UpBound         float64
		LowBoundPercent float64
		LowBound        float64
		deltaPrice      float64
		deltaQuantity   float64

		// Прогресії, використовувалось тіко для grid_v3
		GetDelta                progressions.DeltaType
		NthTerm                 progressions.NthTermType
		Sum                     progressions.SumType
		FindNthTerm             progressions.FindNthTermType
		FindLengthOfProgression progressions.FindLengthOfProgressionType
		FindProgressionTthTerm  progressions.FindCubicProgressionTthTermType

		// Дані про позицію
		callbackRate float64

		// Дані про стакан
		depth *depth_types.Depth
	}
)

const (
	errorMsg = "Error: %v"
)
