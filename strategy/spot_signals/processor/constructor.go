package processor

import (
	"math"
	"time"

	"github.com/adshao/go-binance/v2"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

func NewPairProcessor(
	client *binance.Client,
	symbol string,
	limitOnPosition float64,
	limitOnTransaction float64,
	UpBound float64,
	LowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	callbackRate float64,
	stop chan struct{},
	functions ...Functions) (pp *PairProcessor, err error) {
	exchangeInfo := exchange_types.New()
	err = spot_exchange_info.Init(exchangeInfo, 3, client, symbol)
	if err != nil {
		return
	}
	pp = &PairProcessor{
		client:       client,
		exchangeInfo: exchangeInfo,

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		stop: stop,

		pairInfo:     nil,
		orderTypes:   map[string]bool{},
		degree:       3,
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}
	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.SpotSymbol{Symbol: symbol}).(*symbol_types.SpotSymbol)

	// Ініціалізуємо типи ордерів які можна використовувати для пари
	pp.orderTypes = make(map[string]bool, 0)
	for _, orderType := range pp.pairInfo.OrderTypes {
		pp.orderTypes[orderType] = true
	}

	// Буферизуємо інформацію про символ
	pp.symbol, err = pp.GetSymbol().GetSpotSymbol()
	if err != nil {
		return
	}
	pp.baseSymbol = pp.symbol.QuoteAsset
	pp.targetSymbol = pp.symbol.BaseAsset
	pp.notional = utils.ConvStrToFloat64(pp.symbol.NotionalFilter().MinNotional)
	pp.stepSizeDelta = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)

	// Перевіряємо ліміти на ордери та запити
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit =
		LimitRead(pp.degree, []string{pp.symbol.Symbol}, client)

	if functions != nil {
		pp.testUp = functions[0].TestUp
		pp.testDown = functions[0].TestDown
		pp.NextPriceUp = functions[0].NextPriceUp
		pp.NextPriceDown = functions[0].NextPriceDown
		pp.NextQuantityUp = functions[0].NextQuantityUp
		pp.NextQuantityDown = functions[0].NextQuantityDown
	} else {
		pp.testUp = func(s, e float64) bool { return s < e }
		pp.testDown = func(s, e float64) bool { return s > e }
		pp.NextPriceUp = func(s float64, n int) float64 {
			return pp.roundPrice(s * math.Pow(1+deltaPrice, float64(2)))
		}
		pp.NextPriceDown = func(s float64, n int) float64 {
			return pp.roundPrice(s * math.Pow(1-deltaPrice, float64(2)))
		}
		pp.NextQuantityUp = func(s float64, n int) float64 {
			return pp.roundQuantity(s * (math.Pow(1+deltaQuantity, float64(2))))
		}
		pp.NextQuantityDown = func(s float64, n int) float64 {
			return pp.roundQuantity(s * (math.Pow(1+deltaQuantity, float64(2))))
		}
	}

	return
}
