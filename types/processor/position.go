package processor

import (
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

// func (pp *PairProcessor) getPositionRisk(times int) (risks []*futures.PositionRisk, err error) {
// 	if times == 0 {
// 		return
// 	}
// 	risks, err = pp.client.NewGetPositionRiskService().Symbol(pp.pairInfo.GetSymbol()).Do(context.Background())
// 	if err != nil {
// 		errApi, _ := utils.ParseAPIError(err)
// 		if errApi != nil && errApi.Code == -1021 {
// 			time.Sleep(3 * time.Second)
// 			return pp.getPositionRisk(times - 1)
// 		}
// 	}
// 	return
// }

func (pp *Processor) GetPositionRisk() (risk *futures.PositionRisk, err error) {
	if pp.getPositionRisk == nil {
		return nil, fmt.Errorf("getPositionRisk is not set")
	}
	return pp.getPositionRisk()
}

func (pp *Processor) GetLiquidationDistance(price float64) (distance float64) {
	risk, _ := pp.GetPositionRisk()
	return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
}

func (pp *Processor) GetLeverage() int {
	if pp.getLeverage == nil {
		return 0
	}
	return pp.getLeverage()
}

func (pp *Processor) SetLeverage(leverage int) (res *futures.SymbolLeverage, err error) {
	if pp.setLeverage == nil {
		return nil, fmt.Errorf("setLeverage is not set")
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
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	margin = utils.ConvStrToFloat64(risk.IsolatedMargin) // Convert string to float64
	return
}

func (pp *Processor) SetPositionMargin(amountMargin items_types.ValueType, typeMargin int) (err error) {
	if pp.setPositionMargin == nil {
		return fmt.Errorf("setPositionMargin is not set")
	}
	return pp.setPositionMargin(amountMargin, typeMargin)
}

func (pp *Processor) ClosePosition(risk *futures.PositionRisk) (err error) {
	if pp.closePosition == nil {
		return fmt.Errorf("closePosition is not set")
	}
	return pp.closePosition(risk)
}

func (pp *Processor) GetPositionAmt() (positionAmt float64) {
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	positionAmt = utils.ConvStrToFloat64(risk.PositionAmt) // Convert string to float64
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
