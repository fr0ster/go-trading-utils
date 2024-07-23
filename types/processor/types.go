package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

// const (
// 	errorMsg    = "Error: %v"
// 	repeatTimes = 3
// )

type (
	// Дані про обмеження на пару
	SymbolInfo struct {
		notional items_types.ValueType
		StepSize items_types.QuantityType
		maxQty   items_types.QuantityType
		minQty   items_types.QuantityType
		tickSize items_types.PriceType
		maxPrice items_types.PriceType
		minPrice items_types.PriceType
	}
	Processor struct {
		// Дані про клієнта
		// client *futures.Client
		// Налаштування та обмеження, реалізація
		orderTypes map[futures.OrderType]bool
		degree     int
		timeOut    time.Duration

		// Дані про біржу
		exchangeInfo *exchange_types.ExchangeInfo

		// Дані про пару
		// symbol       *futures.Symbol
		// pairInfo     *symbol_types.FuturesSymbol
		symbol string
		// baseSymbol   string
		// targetSymbol string

		// Дані про обмеження на пару
		symbolInfo SymbolInfo

		// канал зупинки
		stop chan struct{}

		// Дані про стакан
		depth  *depth_types.Depths
		orders *orders_types.Orders

		GetBaseBalance   func() (items_types.ValueType, error)
		GetTargetBalance func() (items_types.ValueType, error)
		GetFreeBalance   func() items_types.ValueType
		GetLockedBalance func() (items_types.ValueType, error)
		GetCurrentPrice  func() (items_types.PriceType, error)
		GetSymbolInfo    func() (SymbolInfo, error)

		getPositionRisk func() (risks *futures.PositionRisk, err error)

		setLeverage func(leverage int) (res *futures.SymbolLeverage, err error)
		getLeverage func() int

		setMarginType func(marginType pairs_types.MarginType) (err error)
		getMarginType func() pairs_types.MarginType

		setPositionMargin func(amountMargin items_types.ValueType, typeMargin int) (err error)

		getCallbackRate func() items_types.PricePercentType

		closePosition func(risk *futures.PositionRisk) (err error)

		getDeltaPrice         func() items_types.PriceType
		getDeltaQuantity      func() items_types.QuantityType
		getLimitOnTransaction func() (limit items_types.ValueType)
		getUpAndLowBound      func() items_types.PricePercentType
	}
)
