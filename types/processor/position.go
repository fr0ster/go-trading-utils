package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
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
func (pp *Processor) GetMarginType() pairs_types.MarginType {
	if pp.getMarginType == nil {
		return ""
	}
	return pp.getMarginType()
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *Processor) SetMarginType(marginType pairs_types.MarginType) (err error) {
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

func (pp *Processor) GetPredictableLoss(risk *futures.PositionRisk, price items_types.PriceType) (unRealizedProfit items_types.ValueType) {
	if risk == nil || pp.GetLeverage() <= 0 {
		return 0
	}
	entryPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
	positionAmt := items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	if positionAmt == 0 { // No position
		return 0
	} else if positionAmt < 0 { // Short position
		unRealizedProfit = items_types.ValueType(float64(pp.GetUpBound(price)-entryPrice) * float64(positionAmt))
	} else if positionAmt > 0 { // Long position
		unRealizedProfit = items_types.ValueType(float64(entryPrice-pp.GetLowBound(price)) * float64(positionAmt))
	}
	return
}
func (pp *Processor) CheckAddPosition(risk *futures.PositionRisk, price items_types.PriceType) bool {
	if risk == nil {
		return false
	}
	positionAmt := items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	liquidationPrice := items_types.PriceType(utils.ConvStrToFloat64(risk.LiquidationPrice))
	if positionAmt == 0 { // No position
		return true
	} else if positionAmt < 0 { // Short position
		return liquidationPrice > pp.GetUpBound(price) &&
			pp.GetPredictableLoss(risk, pp.GetUpBound(price)) > -(pp.GetFreeBalance()*items_types.ValueType(pp.GetLeverage())) &&
			price <= pp.GetUpBound(price)
	} else if positionAmt > 0 { // Long position
		return liquidationPrice < pp.GetLowBound(price) &&
			pp.GetPredictableLoss(risk, pp.GetLowBound(price)) > -(pp.GetFreeBalance()*items_types.ValueType(pp.GetLeverage())) &&
			price >= pp.GetLowBound(price)
	}
	return false
}

func (pp *Processor) CheckStopLoss(free items_types.ValueType, risk *futures.PositionRisk, price items_types.PriceType) bool {
	if risk == nil || utils.ConvStrToFloat64(risk.PositionAmt) == 0 {
		return false
	}
	return (utils.ConvStrToFloat64(risk.PositionAmt) > 0 && price < pp.GetLowBound(price)) ||
		(utils.ConvStrToFloat64(risk.PositionAmt) < 0 && price > pp.GetUpBound(price)) ||
		items_types.ValueType(math.Abs(utils.ConvStrToFloat64(risk.UnRealizedProfit))) > free
}
