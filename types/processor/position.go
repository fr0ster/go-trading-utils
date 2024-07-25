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

func (pp *Processor) GetPositionAmt() (positionAmt items_types.QuantityType) {
	if risk := pp.GetPositionRisk(); risk != nil {
		positionAmt = items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	}
	return
}

func (pp *Processor) GetPredictableProfitOrLoss(positionAmt items_types.QuantityType, firstPrice, secondPrice items_types.PriceType) (unRealizedProfit items_types.ValueType) {
	unRealizedProfit = items_types.ValueType(math.Abs(float64(secondPrice-firstPrice))) * items_types.ValueType(positionAmt)
	return
}
