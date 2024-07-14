package processor

import (
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

func (pp *PairProcessor) RoundPrice(price float64) float64 {
	return utils.RoundToDecimalPlace(price, pp.GetTickSizeExp())
}

func (pp *PairProcessor) RoundQuantity(quantity float64) float64 {
	return utils.RoundToDecimalPlace(quantity, pp.GetStepSizeExp())
}

func (pp *PairProcessor) Debug(fl, id string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		orders, _ := pp.GetOpenOrders()
		logrus.Debugf("%s %s %s:", fl, id, pp.symbol.Symbol)
		for _, order := range orders {
			logrus.Debugf(" Open Order %v on price %v OrderSide %v Status %s", order.OrderID, order.Price, order.Side, order.Status)
		}
	}
}

func (pp *PairProcessor) GetTargetPrices() (priceUp, priceDown float64, err error) {
	if pp.depth != nil {
		priceUp, priceDown = pp.depth.GetTargetPrices(pp.depth.GetPercentToTarget())
	} else {
		err = fmt.Errorf("depth is nil")
	}
	return
}

func (pp *PairProcessor) GetLimitPrices() (priceUp, priceDown float64, err error) {
	if pp.depth != nil {
		priceUp, priceDown = pp.depth.GetTargetPrices(pp.depth.GetPercentToLimit())
	} else {
		err = fmt.Errorf("depth is nil")
	}
	return
}

func (pp *PairProcessor) GetPrices(
	price float64,
	risk *futures.PositionRisk,
	isDynamic bool) (
	priceUp,
	quantityUp,
	priceDown,
	quantityDown float64,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	if pp.depth != nil {
		priceUp, priceDown = pp.depth.GetTargetPrices(pp.depth.GetPercentToTarget())
	} else {
		priceUp = pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
		priceDown = pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
	}
	reduceOnlyUp = false
	reduceOnlyDown = false
	if isDynamic {
		_, _, _, quantityUp, _, err = pp.CalculateInitialPosition(priceUp, pp.UpBound)
		if err != nil {
			quantityUp = 0
		}
		_, _, _, quantityDown, _, err = pp.CalculateInitialPosition(priceDown, pp.LowBound)
		if err != nil {
			quantityDown = 0
		}
	} else {
		quantityUp = pp.RoundQuantity(pp.GetLimitOnTransaction() * float64(pp.GetLeverage()) / priceUp)
		quantityDown = pp.RoundQuantity(pp.GetLimitOnTransaction() * float64(pp.GetLeverage()) / priceDown)
	}
	if quantityUp == 0 && quantityDown == 0 {
		err = fmt.Errorf("can't calculate initial position for price up %v and price down %v", priceUp, priceDown)
		return
	}
	if quantityUp*priceUp < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity up %v * price up %v < notional %v", quantityUp, priceUp, pp.GetNotional())
		return
	} else if quantityDown*priceDown < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity down %v * price down %v < notional %v", quantityDown, priceDown, pp.GetNotional())
		return
	}
	if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		positionPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
		if positionPrice == 0 {
			positionPrice = utils.ConvStrToFloat64(risk.EntryPrice)
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			priceDown = pp.NextPriceDown(math.Min(positionPrice, price))
			quantityDown = -utils.ConvStrToFloat64(risk.PositionAmt)
			reduceOnlyDown = true
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			priceUp = pp.NextPriceUp(math.Max(positionPrice, price))
			quantityUp = utils.ConvStrToFloat64(risk.PositionAmt)
			reduceOnlyUp = true
		}
	}
	return
}

func (pp *PairProcessor) GetTPAndSLOrdersSideAndTypes(
	risk *futures.PositionRisk,
	upOrderSideOpen futures.SideType,
	upPositionNewOrderType futures.OrderType,
	downOrderSideOpen futures.SideType,
	downPositionNewOrderType futures.OrderType,
	shortPositionTPOrderType futures.OrderType,
	shortPositionSLOrderType futures.OrderType,
	longPositionTPOrderType futures.OrderType,
	longPositionSLOrderType futures.OrderType,
	isDynamic bool) (
	upOrderSide futures.SideType,
	upOrderType futures.OrderType,
	downOrderSide futures.SideType,
	downOrderType futures.OrderType,
	err error) {
	upOrderSide = upOrderSideOpen
	upOrderType = upPositionNewOrderType
	downOrderSide = downOrderSideOpen
	downOrderType = downPositionNewOrderType
	if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 &&
		math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)) > pp.GetNotional() {
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 { // SHORT Закриваємо SHORT позицію
			upOrderSide = futures.SideTypeBuy
			upOrderType = shortPositionSLOrderType
			downOrderSide = futures.SideTypeBuy
			downOrderType = shortPositionTPOrderType
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 { // LONG Закриваємо LONG позицію
			upOrderSide = futures.SideTypeSell
			upOrderType = longPositionTPOrderType
			downOrderSide = futures.SideTypeSell
			downOrderType = longPositionSLOrderType
		}
	}
	return
}

func (pp *PairProcessor) LimitRead() (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	err error) {
	exchangeInfo := exchange_types.New()
	futures_exchange_info.RestrictedInit(exchangeInfo, pp.degree, []string{pp.symbol.Symbol}, pp.client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	if minuteRawRequestLimit == nil {
		err = fmt.Errorf("minute raw request limit is not found")
		return
	}
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
