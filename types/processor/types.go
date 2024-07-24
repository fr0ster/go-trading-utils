package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/fr0ster/go-trading-utils/types"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
)

type (
	GetBaseBalanceFunction   func() items_types.ValueType
	GetTargetBalanceFunction func() items_types.ValueType
	GetFreeBalanceFunction   func() items_types.ValueType
	GetLockedBalanceFunction func() items_types.ValueType
	GetCurrentPriceFunction  func() items_types.PriceType

	GetPositionRiskFunction func() *futures.PositionRisk

	SetLeverageFunction       func(leverage int) (Leverage int, MaxNotionalValue string, Symbol string, err error)
	GetLeverageFunction       func() int
	SetMarginTypeFunction     func(types.MarginType) error
	GetMarginTypeFunction     func() types.MarginType
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

		setLeverage SetLeverageFunction
		getLeverage GetLeverageFunction

		setMarginType SetMarginTypeFunction
		getMarginType GetMarginTypeFunction

		setPositionMargin SetPositionMarginFunction

		closePosition ClosePositionFunction

		getDeltaPrice         GetDeltaPriceFunction
		getDeltaQuantity      GetDeltaQuantityFunction
		getLimitOnTransaction GetLimitOnTransactionFunction
		getUpAndLowBound      GetUpAndLowBoundFunction

		getCallbackRate GetCallbackRateFunction
	}
)
