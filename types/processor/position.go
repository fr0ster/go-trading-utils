package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
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
	delta items_types.PriceType,
	leverage int) (minQuantity items_types.QuantityType) {
	minQuantity = pp.FloorQuantity(items_types.QuantityType(value) / (items_types.QuantityType(delta) * items_types.QuantityType(leverage)))
	return
}

func (pp *Processor) PossibleLoss(
	quantity items_types.QuantityType,
	delta items_types.PriceType,
	leverage int) (possibleLoss items_types.ValueType) {
	possibleLoss = items_types.ValueType(delta) * items_types.ValueType(quantity) * items_types.ValueType(leverage)
	return
}

func (pp *Processor) CalcDeltaOnQuantity(limitLossOnTransaction items_types.ValueType, leverage int) (res items_types.PriceOnQuantityType) {
	res = items_types.PriceOnQuantityType(float64(limitLossOnTransaction) / float64(leverage))
	return
}

func (pp *Processor) CalcDeltaPercentOnQuantity(leverage int) (res items_types.PricePercentOnQuantityType) {
	res = items_types.PricePercentOnQuantityType(100 / float64(leverage))
	return
}

func (pp *Processor) CalcQuantityByUPnL(
	price items_types.PriceType,
	debug ...*futures.PositionRisk) (quantity items_types.QuantityType, err error) {
	var (
		oldQuantity     items_types.QuantityType
		oldPossibleLoss items_types.ValueType
		leverage        int
	)
	risk := pp.GetPositionRisk(debug...)
	limitOfPositionLoss := pp.GetLimitOnPosition()
	limitOfTransactionLoss := pp.GetLimitOnTransaction()
	notional := pp.GetNotional()
	if limitOfTransactionLoss < notional {
		err = fmt.Errorf("limit on transaction %f isn't enough for open position with notional %f", limitOfTransactionLoss, notional)
		return
	}
	delta := items_types.PriceType(pp.DeltaLiquidation(pp.GetLeverage())) * price / 100

	oldQuantity = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	leverage = pp.GetLeverage()

	if oldQuantity != 0 {
		oldPossibleLoss = pp.PossibleLoss(oldQuantity, delta, leverage) + items_types.ValueType(utils.ConvStrToFloat64(risk.UnRealizedProfit))
	}

	if oldPossibleLoss > 0 && limitOfPositionLoss-oldPossibleLoss < notional {
		err = fmt.Errorf("we have open position with possible loss %f with price %f and we couldn't open new position with possible loss %f, we need limit of possible loss more than %f",
			oldPossibleLoss,
			price,
			limitOfPositionLoss-oldPossibleLoss,
			notional+oldPossibleLoss)
		return
	}

	deltaOnQuantity := pp.CalcDeltaOnQuantity(limitOfTransactionLoss, leverage)

	quantity = pp.FloorQuantity(items_types.QuantityType(deltaOnQuantity) / items_types.QuantityType(delta))
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
		if profitOrLoss > targetOfLoss {
			err = fmt.Errorf("profit or loss %f is more than limit of loss %f", profitOrLoss, targetOfLoss)
			return
		}
		liquidationPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.LiquidationPrice))
		entryPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
		delta := price*items_types.PriceType(pp.GetUpAndLowBound()/100) + items_types.PriceType(math.Abs(float64(entryPrice-price)))
		if position < 0 && liquidationPrice < price+delta { // Short position
			err = fmt.Errorf("liquidation price %f is less than price %f + delta %f == %f", liquidationPrice, price, delta, price+delta)
		} else if position > 0 && liquidationPrice > price-delta { // Long position
			err = fmt.Errorf("liquidation price %f is more than price %f - delta %f == %f", liquidationPrice, price, delta, price-delta)
		}
	}
	return
}
