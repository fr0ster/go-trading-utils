package processor

import (
	"time"

	"github.com/adshao/go-binance/v2"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
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
		pairInfo     *symbol_types.SymbolInfo
		baseSymbol   symbol_types.QuoteAsset
		targetSymbol symbol_types.BaseAsset

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

		limitOnPosition    float64
		limitOnTransaction float64

		// Дінаміка ціни, використовувалось тіко для grid_v3
		UpBoundPercent  items_types.PricePercentType
		UpBound         items_types.PriceType
		LowBoundPercent items_types.PricePercentType
		LowBound        items_types.PriceType
		deltaPrice      items_types.PriceType
		deltaQuantity   items_types.QuantityType

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
		depth *depth_types.Depths
	}
)

const (
	errorMsg = "Error: %v"
)
