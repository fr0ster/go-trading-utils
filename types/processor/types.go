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

type (
	Processor struct {
		// Налаштування та обмеження, реалізація
		orderTypes map[futures.OrderType]bool
		degree     int
		timeOut    time.Duration

		symbol string

		// Дані про біржу
		exchangeInfo *exchange_types.ExchangeInfo

		// Дані про пару
		// symbolInfo *symbol_types.SymbolInfo

		// канал зупинки
		stop chan struct{}

		// Дані про стакан
		depths *depth_types.Depths
		orders *orders_types.Orders

		GetBaseBalance   func() items_types.ValueType
		GetTargetBalance func() items_types.ValueType
		GetFreeBalance   func() items_types.ValueType
		GetLockedBalance func() items_types.ValueType
		GetCurrentPrice  func() items_types.PriceType
		// getSymbolInfo    func() *symbol_types.SymbolInfo

		getPositionRisk func() *futures.PositionRisk

		setLeverage func(leverage int) (*futures.SymbolLeverage, error)
		getLeverage func() int

		setMarginType func(pairs_types.MarginType) error
		getMarginType func() pairs_types.MarginType

		setPositionMargin func(items_types.ValueType, int) error

		getCallbackRate func() items_types.PricePercentType

		closePosition func(*futures.PositionRisk) (err error)

		getDeltaPrice         func() items_types.PriceType
		getDeltaQuantity      func() items_types.QuantityType
		getLimitOnTransaction func() items_types.ValueType
		getUpAndLowBound      func() items_types.PricePercentType
	}
)
