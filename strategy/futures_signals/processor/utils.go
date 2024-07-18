package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

func (pp *PairProcessor) RoundPrice(price types.PriceType) types.PriceType {
	return types.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *PairProcessor) RoundQuantity(quantity types.QuantityType) types.QuantityType {
	return types.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
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

func (pp *PairProcessor) GetTargetPrices() (priceUp, priceDown types.PriceType, err error) {
	if pp.depth != nil {
		priceUp, priceDown, _, _ = pp.depth.GetTargetPrices(pp.depth.GetPercentToTarget())
	} else {
		err = fmt.Errorf("depth is nil")
	}
	return
}

// func (pp *PairProcessor) GetLimitPrices() (priceUp, priceDown types.PriceType, err error) {
// 	var (
// 		askMax *types.DepthItem
// 		bidMax *types.DepthItem
// 	)
// 	if pp.depth != nil {
// 		askMax, err = pp.depth.AskMax()
// 		if err != nil {
// 			return
// 		}
// 		bidMax, err = pp.depth.BidMax()
// 		if err != nil {
// 			return
// 		}
// 		priceUp = askMax.GetPrice()
// 		priceDown = bidMax.GetPrice()
// 	} else {
// 		err = fmt.Errorf("depth is nil")
// 	}
// 	return
// }

func (pp *PairProcessor) GetPrices(
	price types.PriceType,
	risk *futures.PositionRisk,
	isDynamic bool) (
	priceUp types.PriceType,
	quantityUp types.QuantityType,
	priceDown types.PriceType,
	quantityDown types.QuantityType,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	if pp.depth != nil {
		priceUp, priceDown, err = pp.GetTargetPrices()
		if err != nil {
			return
		}
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
		quantityUp = pp.RoundQuantity(types.QuantityType(float64(pp.GetLimitOnTransaction()) * float64(pp.GetLeverage()) / float64(priceUp)))
		quantityDown = pp.RoundQuantity(types.QuantityType(float64(pp.GetLimitOnTransaction()) * float64(pp.GetLeverage()) / float64(priceDown)))
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
		positionPrice := types.PriceType(utils.ConvStrToFloat64(risk.BreakEvenPrice))
		if positionPrice == 0 {
			positionPrice = types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			priceDown = pp.NextPriceDown(types.PriceType(math.Min(float64(positionPrice), float64(price))))
			quantityDown = types.QuantityType(-utils.ConvStrToFloat64(risk.PositionAmt))
			reduceOnlyDown = true
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			priceUp = pp.NextPriceUp(types.PriceType(math.Max(float64(positionPrice), float64(price))))
			quantityUp = types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
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
