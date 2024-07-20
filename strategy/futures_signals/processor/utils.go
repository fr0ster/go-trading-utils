package processor

import (
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	exchange "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) RoundPrice(price items.PriceType) items.PriceType {
	return items.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *PairProcessor) RoundQuantity(quantity items.QuantityType) items.QuantityType {
	return items.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
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

func (pp *PairProcessor) GetNextUpCoefficient() items.PriceType {
	if pp.depth.GetAsks() == nil ||
		pp.depth.GetBids() == nil ||
		pp.depth.GetAsks().GetSummaValue() == 0 ||
		pp.depth.GetBids().GetSummaValue() == 0 {
		return 1
	}
	coefficients := items.PriceType(pp.depth.GetAsks().GetSummaValue() / pp.depth.GetBids().GetSummaValue())
	if coefficients > 1 {
		return 1
	} else {
		return coefficients
	}
}

func (pp *PairProcessor) GetNextDownCoefficient() items.PriceType {
	if pp.depth.GetAsks() == nil ||
		pp.depth.GetBids() == nil ||
		pp.depth.GetAsks().GetSummaValue() == 0 ||
		pp.depth.GetBids().GetSummaValue() == 0 {
		return 1
	}
	coefficients := items.PriceType(pp.depth.GetAsks().GetSummaValue() / pp.depth.GetBids().GetSummaValue())
	if coefficients > 1 {
		return coefficients
	} else {
		return 1
	}
}

func (pp *PairProcessor) GetTargetPrices(price ...items.PriceType) (priceDown, priceUp items.PriceType, err error) {
	var currentPrice items.PriceType
	if len(price) == 0 {
		currentPrice, err = pp.GetCurrentPrice()
		if err != nil {
			return
		}
	} else {
		currentPrice = price[0]
	}
	priceUp = pp.RoundPrice(currentPrice * (1 + pp.GetDeltaPrice()*pp.GetNextUpCoefficient()))
	priceDown = pp.RoundPrice(currentPrice * (1 - pp.GetDeltaPrice()*pp.GetNextDownCoefficient()))
	return
}

func (pp *PairProcessor) GetLimitPrices() (priceTargetDown, priceTargetUp, priceDown, priceUp items.PriceType, err error) {
	var (
		askMax *items.DepthItem
		bidMax *items.DepthItem
	)
	priceTargetUp, priceTargetDown, err = pp.GetTargetPrices()
	if err != nil {
		return
	}
	asksFilter := func(i *items.DepthItem) bool {
		return i.GetPrice() > priceTargetUp
	}
	bidsFilter := func(i *items.DepthItem) bool {
		return i.GetPrice() < priceTargetDown
	}
	_, askMax = pp.depth.GetAsks().GetFiltered(asksFilter).GetMinMaxByValue()
	_, bidMax = pp.depth.GetBids().GetFiltered(bidsFilter).GetMinMaxByValue()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	return
}

func (pp *PairProcessor) GetPrices(
	price items.PriceType,
	risk *futures.PositionRisk,
	isDynamic bool) (
	priceUp items.PriceType,
	quantityUp items.QuantityType,
	priceDown items.PriceType,
	quantityDown items.QuantityType,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	priceDown, priceUp, err = pp.GetTargetPrices()
	if err != nil {
		return
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
		quantityUp = pp.RoundQuantity(items.QuantityType(float64(pp.GetLimitOnTransaction()) * float64(pp.GetLeverage()) / float64(priceUp)))
		quantityDown = pp.RoundQuantity(items.QuantityType(float64(pp.GetLimitOnTransaction()) * float64(pp.GetLeverage()) / float64(priceDown)))
	}
	if quantityUp == 0 && quantityDown == 0 {
		err = fmt.Errorf("can't calculate initial position for price up %v and price down %v", priceUp, priceDown)
		return
	}
	if float64(quantityUp)*float64(priceUp) < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity up %v * price up %v < notional %v", quantityUp, priceUp, pp.GetNotional())
		return
	} else if float64(quantityUp)*float64(priceUp) < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity down %v * price down %v < notional %v", quantityDown, priceDown, pp.GetNotional())
		return
	}
	if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		positionPrice := items.PriceType(utils.ConvStrToFloat64(risk.BreakEvenPrice))
		if positionPrice == 0 {
			positionPrice = items.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			priceDown = pp.NextPriceDown(items.PriceType(math.Min(float64(positionPrice), float64(price))))
			quantityDown = items.QuantityType(-utils.ConvStrToFloat64(risk.PositionAmt))
			reduceOnlyDown = true
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			priceUp = pp.NextPriceUp(items.PriceType(math.Max(float64(positionPrice), float64(price))))
			quantityUp = items.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
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

func LimitRead(degree int, symbols []string, client *futures.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange.RateLimits,
	dayOrderLimit *exchange.RateLimits,
	minuteRawRequestLimit *exchange.RateLimits) {
	exchangeInfo := exchange.New()
	futures_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
