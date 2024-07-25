package processor

import (
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *Processor) RoundValue(value items_types.ValueType) items_types.ValueType {
	return items_types.ValueType(utils.RoundToDecimalPlace(float64(value), pp.GetTickSizeExp()))
}

func (pp *Processor) RoundPrice(price items_types.PriceType) items_types.PriceType {
	return items_types.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *Processor) RoundQuantity(quantity items_types.QuantityType) items_types.QuantityType {
	return items_types.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
}

func (pp *Processor) GetTargetPrices(price ...items_types.PriceType) (priceDown, priceUp items_types.PriceType, err error) {
	priceUp = pp.NextPriceUp(price...)
	priceDown = pp.NextPriceDown(price...)
	return
}

func (pp *Processor) GetLimitPrices(price ...items_types.PriceType) (priceTargetDown, priceTargetUp, priceDown, priceUp items_types.PriceType, err error) {
	var (
		askMax *items_types.DepthItem
		bidMax *items_types.DepthItem
	)
	priceTargetDown, priceTargetUp, err = pp.GetTargetPrices(price...)
	if err != nil {
		return
	}
	asksFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() > priceTargetUp
	}
	bidsFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() < priceTargetDown
	}
	_, askMax = pp.depths.GetAsks().GetFiltered(asksFilter).GetMinMaxByValue()
	_, bidMax = pp.depths.GetBids().GetFiltered(bidsFilter).GetMinMaxByValue()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	return
}

func (pp *Processor) GetPrices(
	price items_types.PriceType,
	risk *futures.PositionRisk) (
	priceUp items_types.PriceType,
	quantityUp items_types.QuantityType,
	priceDown items_types.PriceType,
	quantityDown items_types.QuantityType,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	priceDown, priceUp, err = pp.GetTargetPrices(price)
	if err != nil {
		return
	}
	reduceOnlyUp = false
	reduceOnlyDown = false
	quantityUp = pp.RoundQuantity(items_types.QuantityType(float64(pp.GetLimitOnTransaction()) * float64(pp.GetLeverage()) / float64(priceUp)))
	quantityDown = pp.RoundQuantity(items_types.QuantityType(float64(pp.GetLimitOnTransaction()) * float64(pp.GetLeverage()) / float64(priceDown)))
	if quantityUp == 0 && quantityDown == 0 {
		err = fmt.Errorf("can't calculate initial position for price up %v and price down %v", priceUp, priceDown)
		return
	}
	if items_types.ValueType(float64(quantityUp)*float64(priceUp)) < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity up %v * price up %v < notional %v", quantityUp, priceUp, pp.GetNotional())
		return
	} else if items_types.ValueType(float64(quantityUp)*float64(priceUp)) < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity down %v * price down %v < notional %v", quantityDown, priceDown, pp.GetNotional())
		return
	}
	if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		breakEvenPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.BreakEvenPrice))
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			priceDown = pp.NextPriceDown(breakEvenPrice)
			quantityDown = items_types.QuantityType(-utils.ConvStrToFloat64(risk.PositionAmt))
			reduceOnlyDown = true
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			priceUp = pp.NextPriceUp(breakEvenPrice)
			quantityUp = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
			reduceOnlyUp = true
		}
	}
	return
}

func (pp *Processor) GetTPAndSLOrdersSideAndTypes(
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
		items_types.ValueType(math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))) > pp.GetNotional() {
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

func (pp *Processor) LimitRead() (
	updateTime time.Duration,
	minuteOrderLimit *exchange.RateLimits,
	dayOrderLimit *exchange.RateLimits,
	minuteRawRequestLimit *exchange.RateLimits) {
	minuteOrderLimit = pp.exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = pp.exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = pp.exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
