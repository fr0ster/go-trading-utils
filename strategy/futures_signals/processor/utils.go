package processor

import (
	"context"
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

func (pp *PairProcessor) GetCurrentPrice() (float64, error) {
	price, err := pp.client.NewListPricesService().Symbol(pp.symbol.Symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func (pp *PairProcessor) GetPrices(
	price float64,
	risk *futures.PositionRisk,
	isDynamic bool) (
	priceUp,
	quantityUp,
	priceDown,
	quantityDown float64,
	err error) {
	priceUp = pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
	priceDown = pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
	if isDynamic {
		_, _, _, quantityUp, _, err = pp.CalculateInitialPosition(priceUp, pp.UpBound)
		if err != nil {
			err = fmt.Errorf("can't calculate initial position for price up %v", priceUp)
			return
		}
		_, _, _, quantityDown, _, err = pp.CalculateInitialPosition(priceDown, pp.LowBound)
		if err != nil {
			err = fmt.Errorf("can't calculate initial position for price down %v", priceDown)
			return
		}
	} else {
		quantityUp = pp.RoundQuantity(pp.GetLimitOnTransaction() * float64(pp.GetLeverage()) / priceUp)
		quantityDown = pp.RoundQuantity(pp.GetLimitOnTransaction() * float64(pp.GetLeverage()) / priceDown)
	}
	if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		positionPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
		if positionPrice == 0 {
			positionPrice = utils.ConvStrToFloat64(risk.EntryPrice)
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 && math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)) > pp.GetNotional() {
			priceDown = pp.NextPriceDown(math.Min(positionPrice, price))
			quantityDown = math.Max(-utils.ConvStrToFloat64(risk.PositionAmt), quantityDown)
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 && math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)) > pp.GetNotional() {
			priceUp = pp.NextPriceUp(math.Max(positionPrice, price))
			quantityUp = math.Max(utils.ConvStrToFloat64(risk.PositionAmt), quantityUp)
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
