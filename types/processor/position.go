package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *Processor) GetPositionRisk() *futures.PositionRisk {
	if pp.getPositionRisk != nil {
		return pp.getPositionRisk()
	}
	return nil
}

func (pp *Processor) GetLiquidationDistance(price float64) (distance float64) {
	if risk := pp.GetPositionRisk(); risk != nil {
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

func (pp *Processor) GetPositionMargin() (margin float64) {
	if risk := pp.GetPositionRisk(); risk != nil {
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

func (pp *Processor) GetPositionAmt() (positionAmt float64) {
	if risk := pp.GetPositionRisk(); risk != nil {
		positionAmt = utils.ConvStrToFloat64(risk.PositionAmt)
	}
	return
}

func (pp *Processor) GetPredictableProfitOrLoss(positionAmt items_types.QuantityType, price items_types.PriceType) (unRealizedProfit items_types.ValueType) {
	if positionAmt == 0 { // No position
		return 0
	} else if positionAmt < 0 { // Short position
		unRealizedProfit = items_types.ValueType(float64(pp.GetUpBound(price)-price) * float64(positionAmt))
	} else if positionAmt > 0 { // Long position
		unRealizedProfit = items_types.ValueType(float64(price-pp.GetLowBound(price)) * float64(positionAmt))
	}
	return
}

func (pp *Processor) CheckPosition(positionAmt items_types.QuantityType, liquidationPrice, price items_types.PriceType) bool {
	if positionAmt == 0 { // No position
		return true
	} else {
		profitOrLoss := pp.GetPredictableProfitOrLoss(positionAmt, price)
		free := pp.getLimitOnPosition()
		if positionAmt < 0 { // Short position
			upBound := pp.GetUpBound(price)
			return liquidationPrice > upBound &&
				profitOrLoss > -free
		} else if positionAmt > 0 { // Long position
			lowBound := pp.GetLowBound(price)
			return liquidationPrice < lowBound &&
				profitOrLoss > -free
		}
	}
	return false
}
