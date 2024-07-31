package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *Processor) DeltaLiquidation(leverage int, lossPercent ...items_types.ValuePercentType) (res items_types.PricePercentType) {
	if len(lossPercent) != 0 {
		res = items_types.PricePercentType(float64(lossPercent[0]) * 100 / float64(leverage))
	} else {
		res = items_types.PricePercentType(100 / float64(leverage))
	}
	return
}

func (pp *Processor) PossibleQuantity(
	value items_types.ValueType,
	delta items_types.DeltaPriceType,
	round ...bool) (possibleQuantity items_types.QuantityType) {
	if len(round) > 0 && !round[0] {
		possibleQuantity = items_types.QuantityType(value) / items_types.QuantityType(delta)
	} else {
		possibleQuantity = pp.FloorQuantity(items_types.QuantityType(value) / items_types.QuantityType(delta))
	}
	return
}

func (pp *Processor) PossibleLoss(
	quantity items_types.QuantityType,
	delta items_types.DeltaPriceType) (possibleLoss items_types.ValueType) {
	possibleLoss = items_types.ValueType(quantity) * (items_types.ValueType(delta))
	return
}

func (pp *Processor) CalcQuantityByUPnL(
	upOrDown depth_types.UpOrDown,
	price items_types.PriceType,
	debug ...*futures.PositionRisk) (newQuantity items_types.QuantityType, err error) {
	var (
		position         items_types.QuantityType
		fullPossibleLoss items_types.ValueType
		leverage         int
	)
	risk := pp.GetPositionRisk(debug...)
	limitOfPositionLoss := pp.GetLimitOnPosition()
	// Частка на транзакцію залежить від наявних коштів, бо якшо маємо коштів меньше ліміту на позицію, то і ліміт на транзакцію відповідно менший
	limitOfTransactionLoss := pp.GetLimitOnTransaction()
	notional := pp.GetNotional()
	if limitOfTransactionLoss < notional {
		err = fmt.Errorf("limit on transaction %f isn't enough for open position with notional %f", limitOfTransactionLoss, notional)
		return
	}

	position = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	leverage = pp.GetLeverage()
	deltaLiquidation := pp.DeltaLiquidation(leverage)
	newQuantity = pp.PossibleQuantity(limitOfTransactionLoss, items_types.DeltaPriceType(price)*items_types.DeltaPriceType(deltaLiquidation/100))
	if upOrDown == depth_types.UP && position < 0 || upOrDown == depth_types.DOWN && position > 0 {

		if position != 0 {
			fullPossibleLoss = pp.PossibleLoss(
				items_types.QuantityType(math.Abs(float64(position+newQuantity))),
				items_types.DeltaPriceType(price*items_types.PriceType(deltaLiquidation/100))) -
				items_types.ValueType(utils.ConvStrToFloat64(risk.UnRealizedProfit))
		}

		if fullPossibleLoss > 0 && limitOfPositionLoss-fullPossibleLoss < notional {
			newQuantity = 0
			err = fmt.Errorf("we have open position with possible loss %f with price %f and we couldn't open new position with possible loss %f, we need limit of possible loss more than %f",
				fullPossibleLoss,
				price,
				limitOfPositionLoss-fullPossibleLoss,
				notional+fullPossibleLoss)
			return
		}
	}
	return
}

func (pp *Processor) CheckPosition(
	price items_types.PriceType,
	debug ...*futures.PositionRisk) (err error) {
	var (
		position items_types.QuantityType
	)
	risk := pp.GetPositionRisk(debug...)
	if risk == nil {
		return
	}
	position = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	targetOfLoss := pp.GetLimitOnPosition()
	if position == 0 { // No position
		return
	} else {
		profitOrLoss := items_types.ValueType(utils.ConvStrToFloat64(risk.UnRealizedProfit))
		if profitOrLoss < -targetOfLoss {
			err = fmt.Errorf("profit or loss %f is more than limit of loss %f", profitOrLoss, targetOfLoss)
			return
		}
	}
	return
}
