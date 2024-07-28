package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *Processor) NextPriceUp(prices ...items_types.PriceType) items_types.PriceType {
	price := items_types.PriceType(0.0)
	if len(prices) == 0 {
		price = pp.GetCurrentPrice()
	} else {
		price = prices[0]
	}
	return price * items_types.PriceType(1+pp.GetDeltaPrice()/100)
}

func (pp *Processor) NextPriceDown(prices ...items_types.PriceType) items_types.PriceType {
	price := items_types.PriceType(0.0)
	if len(prices) == 0 {
		price = pp.GetCurrentPrice()
	} else {
		price = prices[0]
	}
	return price * items_types.PriceType(1-pp.GetDeltaPrice()/100)
}

func (pp *Processor) NextQuantityUp(quantity items_types.QuantityType) items_types.QuantityType {
	return quantity * items_types.QuantityType(1+pp.GetDeltaQuantity()/100)
}

func (pp *Processor) NextQuantityDown(quantity items_types.QuantityType) items_types.QuantityType {
	return quantity * items_types.QuantityType(1-pp.GetDeltaQuantity())
}

func (pp *Processor) GetPredictableProfitOrLoss(
	quantity items_types.QuantityType,
	delta items_types.PriceType) (unRealizedProfit items_types.ValueType) {
	unRealizedProfit = items_types.ValueType(delta) * items_types.ValueType(quantity) * items_types.ValueType(pp.GetLeverage())
	return
}

func (pp *Processor) MinPossibleLoss(
	price items_types.PriceType,
	delta items_types.PriceType,
	leverage int) (minQuantity items_types.QuantityType, minPossibleLoss items_types.ValueType) {
	notional := pp.GetNotional()

	minQuantity = pp.CeilQuantity(items_types.QuantityType(notional) / items_types.QuantityType(price))
	minPossibleLoss = items_types.ValueType(delta) * items_types.ValueType(minQuantity) * items_types.ValueType(leverage)
	return
}

func (pp *Processor) CalcQuantityByUPnL(
	price items_types.PriceType,
	delta items_types.PriceType,
	isCorrected bool,
	debug ...*futures.PositionRisk) (quantity items_types.QuantityType, err error) {
	var (
		oldQuantity     items_types.QuantityType
		oldDelta        items_types.PriceType
		oldPossibleLoss items_types.ValueType
		leverage        int
	)
	risk := pp.GetPositionRisk(debug...)
	oldQuantity = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	leverage = pp.GetLeverage()
	targetOfPossibleLoss := pp.GetLimitOnPosition()
	transaction := pp.GetLimitOnTransaction()

	if oldQuantity != 0 {
		oldDelta = items_types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice)-float64(price)) + delta
		oldPossibleLoss = items_types.ValueType(oldDelta) * items_types.ValueType(oldQuantity) * items_types.ValueType(leverage)
	}

	minQuantity, minLoss := pp.MinPossibleLoss(price, delta, leverage)

	if transaction < minLoss {
		err = fmt.Errorf("limit on transaction %f with price %f isn't enough for open position with leverage %d, we need at least %f or decrease leverage",
			transaction, price, leverage, minLoss)
		return
	}

	if targetOfPossibleLoss-oldPossibleLoss < minLoss {
		if oldPossibleLoss > 0 {
			err = fmt.Errorf("we have open position with possible loss %f with price %f and we couldn't open new position with possible loss %f, we need limit of possible loss more than %f",
				oldPossibleLoss,
				price,
				targetOfPossibleLoss-oldPossibleLoss,
				minLoss+oldPossibleLoss)
		} else {
			err = fmt.Errorf("target of loss %f with price %f is less than min loss %f", targetOfPossibleLoss, price, minLoss)
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
			err = fmt.Errorf("limit on transaction %f isn't enough for open position with leverage %d, we need at least %f or decrease leverage",
				transaction, leverage, minLoss)
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
