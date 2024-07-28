package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/fr0ster/go-trading-utils/types"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	DepthConstructor  func() *depth_types.Depths
	OrdersConstructor func() *orders_types.Orders

	GetBaseBalanceFunction   func() items_types.ValueType
	GetTargetBalanceFunction func() items_types.QuantityType
	GetFreeBalanceFunction   func() items_types.ValueType
	GetLockedBalanceFunction func() items_types.ValueType
	GetCurrentPriceFunction  func() items_types.PriceType

	GetPositionRiskFunction func() *futures.PositionRisk

	SetLeverageFunction func(leverage int) (Leverage int, MaxNotionalValue string, Symbol string, err error)
	GetLeverageFunction func() int

	GetMarginTypeFunction     func() types.MarginType
	SetMarginTypeFunction     func(types.MarginType) error
	SetPositionMarginFunction func(items_types.ValueType, int) error

	ClosePositionFunction func() (err error)

	GetDeltaPriceFunction    func() items_types.PricePercentType
	GetDeltaQuantityFunction func() items_types.QuantityPercentType

	GetLimitOnPositionFunction    func() items_types.ValueType
	GetLimitOnTransactionFunction func() items_types.ValuePercentType

	GetUpAndLowBoundFunction func() items_types.PricePercentType

	GetCallbackRateFunction func() items_types.PricePercentType
	Processor               struct {
		// Налаштування та обмеження, реалізація
		orderTypes map[futures.OrderType]bool
		degree     int
		timeOut    time.Duration

		symbol string

		// Дані про біржу
		exchangeInfo *exchange_types.ExchangeInfo
		symbolInfo   *symbol_types.Symbol

		// канал зупинки
		stop chan struct{}

		// Дані про стакан
		depths *depth_types.Depths
		orders *orders_types.Orders

		getBaseBalance   GetBaseBalanceFunction
		getTargetBalance GetTargetBalanceFunction
		getFreeBalance   GetFreeBalanceFunction
		getLockedBalance GetLockedBalanceFunction
		getCurrentPrice  GetCurrentPriceFunction

		getPositionRisk GetPositionRiskFunction

		getLeverage GetLeverageFunction
		setLeverage SetLeverageFunction

		setMarginType SetMarginTypeFunction
		getMarginType GetMarginTypeFunction

		setPositionMargin SetPositionMarginFunction

		closePosition ClosePositionFunction

		getDeltaPrice         GetDeltaPriceFunction
		getDeltaQuantity      GetDeltaQuantityFunction
		getLimitOnPosition    GetLimitOnPositionFunction
		getLimitOnTransaction GetLimitOnTransactionFunction
		getUpAndLowBound      GetUpAndLowBoundFunction

		getCallbackRate GetCallbackRateFunction
	}
)
