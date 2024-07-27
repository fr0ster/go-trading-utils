package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *Processor) GetPositionRisk(debug ...*futures.PositionRisk) *futures.PositionRisk {
	if len(debug) > 0 {
		return debug[0]
	}
	if pp.getPositionRisk != nil {
		return pp.getPositionRisk()
	}
	return nil
}

func (pp *Processor) GetLiquidationDistance(price float64, debug ...*futures.PositionRisk) (distance float64) {
	if risk := pp.GetPositionRisk(debug...); risk != nil {
		return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
	} else {
		return 0
	}
}

func (pp *Processor) GetLeverage() int {
	if pp.getLeverage == nil {
		return 0
	}
	return pp.getLeverage()
}

func (pp *Processor) SetLeverage(leverage int) (Leverage int, MaxNotionalValue string, Symbol string, err error) {
	if pp.setLeverage == nil {
		err = fmt.Errorf("setLeverage is not set")
		return
	}
	return pp.setLeverage(leverage)
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *Processor) GetMarginType() types.MarginType {
	if pp.getMarginType == nil {
		return ""
	}
	return pp.getMarginType()
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *Processor) SetMarginType(marginType types.MarginType) (err error) {
	if pp.setMarginType == nil {
		return fmt.Errorf("setMarginType is not set")
	}
	return pp.setMarginType(marginType)
}

func (pp *Processor) GetPositionMargin(debug ...*futures.PositionRisk) (margin float64) {
	if risk := pp.GetPositionRisk(debug...); risk != nil {
		margin = utils.ConvStrToFloat64(risk.IsolatedMargin) // Convert string to float64
	}
	return
}

func (pp *Processor) SetPositionMargin(amountMargin items_types.ValueType, typeMargin int) (err error) {
	if pp.setPositionMargin == nil {
		return fmt.Errorf("setPositionMargin is not set")
	}
	return pp.setPositionMargin(amountMargin, typeMargin)
}

func (pp *Processor) ClosePosition() (err error) {
	if pp.closePosition == nil {
		return fmt.Errorf("closePosition is not set")
	}
	return pp.closePosition()
}

func (pp *Processor) GetPositionAmt(debug ...*futures.PositionRisk) (positionAmt items_types.QuantityType) {
	if risk := pp.GetPositionRisk(debug...); risk != nil {
		positionAmt = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	}
	return
}

func (pp *Processor) GetPredictableProfitOrLoss(
	quantity items_types.QuantityType,
	delta items_types.PriceType) (unRealizedProfit items_types.ValueType) {
	unRealizedProfit = items_types.ValueType(delta) * items_types.ValueType(quantity) * items_types.ValueType(pp.GetLeverage())
	return
}

func (pp *Processor) GetQuantityByUPnL(
	price items_types.PriceType,
	delta items_types.PriceType,
	isCorrected bool,
	debug ...*futures.PositionRisk) (quantity items_types.QuantityType, err error) {
	var (
		oldQuantity     items_types.QuantityType
		oldDelta        items_types.PriceType
		oldPossibleLoss items_types.ValueType
	)
	risk := pp.GetPositionRisk(debug...)
	notional := items_types.ValueType(utils.ConvStrToFloat64(risk.Notional))
	if notional == 0 {
		notional = pp.GetNotional()
	}
	leverage := int(utils.ConvStrToFloat64(risk.Leverage))
	if leverage == 0 {
		leverage = pp.GetLeverage()
	}
	targetOfPossibleLoss := pp.GetLimitOnPosition()
	transaction := pp.GetLimitOnTransaction()
	if transaction < notional {
		err = fmt.Errorf("limit on transaction %f isn't enough for open position with leverage %d, we need at least %f",
			transaction, leverage, notional)
		return
	}

	oldQuantity = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	if oldQuantity != 0 {
		oldDelta = items_types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice)-float64(price)) + delta
		oldPossibleLoss = items_types.ValueType(oldDelta) * items_types.ValueType(oldQuantity) * items_types.ValueType(leverage)
	}

	minQuantity := pp.CeilQuantity(items_types.QuantityType(notional) / items_types.QuantityType(price))
	minLoss := items_types.ValueType(delta) * items_types.ValueType(minQuantity) * items_types.ValueType(leverage)

	if targetOfPossibleLoss-oldPossibleLoss < minLoss {
		if oldPossibleLoss > 0 {
			err = fmt.Errorf("we have open position with possible loss %f and we couldn't open new position with possible loss %f, we need limit of possible loss more than %f",
				oldPossibleLoss,
				targetOfPossibleLoss-oldPossibleLoss,
				minLoss+oldPossibleLoss)
		} else {
			err = fmt.Errorf("target of loss %f is less than min loss %f", targetOfPossibleLoss, minLoss)
		}
		if isCorrected {
			quantity = minQuantity
			err = nil
		}
		return
	}

	deltaOnQuantity := transaction / items_types.ValueType(leverage)

	quantity = pp.FloorQuantity(items_types.QuantityType(deltaOnQuantity) / items_types.QuantityType(delta))
	if quantity < minQuantity {
		if isCorrected {
			quantity = minQuantity
			err = nil
		} else {
			err = fmt.Errorf("limit on transaction %f isn't enough for open position with leverage %d, we need at least %f",
				transaction, leverage, notional)
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
		if profitOrLoss > targetOfLoss {
			err = fmt.Errorf("profit or loss %f is more than limit of loss %f", profitOrLoss, targetOfLoss)
			return
		}
		liquidationPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.LiquidationPrice))
		entryPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
		delta := price*items_types.PriceType(pp.GetUpAndLowBound()/100) + items_types.PriceType(math.Abs(float64(entryPrice-price)))
		if position < 0 && // Short position
			liquidationPrice < price+delta {
			err = fmt.Errorf("liquidation price %f is less than price %f + delta %f == %f", liquidationPrice, price, delta, price+delta)
		} else if position > 0 && // Long position
			liquidationPrice > price-delta {
			err = fmt.Errorf("liquidation price %f is more than price %f - delta %f == %f", liquidationPrice, price, delta, price-delta)
		}
	}
	return
}
