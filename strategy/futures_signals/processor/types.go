package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/google/btree"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
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
		notional items_types.ValueType
		StepSize items_types.QuantityType
		maxQty   items_types.QuantityType
		minQty   items_types.QuantityType
		tickSize items_types.PriceType
		maxPrice items_types.PriceType
		minPrice items_types.PriceType

		// канал зупинки
		stop chan struct{}

		limitOnPosition    items_types.ValueType
		limitOnTransaction items_types.ValuePercentType

		// Дінаміка ціни, використовувалось тіко для grid_v3
		minSteps        int
		up              *btree.BTree
		down            *btree.BTree
		UpBoundPercent  items_types.PricePercentType
		UpBound         items_types.PriceType
		LowBoundPercent items_types.PricePercentType
		LowBound        items_types.PriceType
		deltaPrice      items_types.PriceType
		deltaQuantity   items_types.QuantityType

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
		callbackRate items_types.PricePercentType

		// Дані про стакан
		depth *depth_types.Depths
	}
)
