package processor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) getPositionRisk(times int) (risks []*futures.PositionRisk, err error) {
	if times == 0 {
		return
	}
	risks, err = pp.client.NewGetPositionRiskService().Symbol(pp.pairInfo.GetSymbol()).Do(context.Background())
	if err != nil {
		errApi, _ := utils.ParseAPIError(err)
		if errApi != nil && errApi.Code == -1021 {
			time.Sleep(3 * time.Second)
			return pp.getPositionRisk(times - 1)
		}
	}
	return
}

func (pp *PairProcessor) GetPositionRisk() (risks *futures.PositionRisk, err error) {
	risk, err := pp.getPositionRisk(3)
	if err != nil {
		return nil, err
	} else if len(risk) == 0 {
		return nil, fmt.Errorf("can't get position risk for symbol %s", pp.symbol.Symbol)
	} else {
		return risk[0], nil
	}
}

func (pp *PairProcessor) GetLiquidationDistance(price float64) (distance float64) {
	risk, _ := pp.GetPositionRisk()
	return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
}

func (pp *PairProcessor) GetLeverage() int {
	return pp.leverage
}

func (pp *PairProcessor) SetLeverage(leverage int) (res *futures.SymbolLeverage, err error) {
	return pp.client.NewChangeLeverageService().Symbol(pp.symbol.Symbol).Leverage(leverage).Do(context.Background())
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) GetMarginType() pairs_types.MarginType {
	return pp.marginType
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) SetMarginType(marginType pairs_types.MarginType) (err error) {
	return pp.client.
		NewChangeMarginTypeService().
		Symbol(pp.symbol.Symbol).
		MarginType(futures.MarginType(marginType)).
		Do(context.Background())
}

func (pp *PairProcessor) GetPositionMargin() (margin float64) {
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	margin = utils.ConvStrToFloat64(risk.IsolatedMargin) // Convert string to float64
	return
}

func (pp *PairProcessor) SetPositionMargin(amountMargin types.PriceType, typeMargin int) (err error) {
	return pp.client.NewUpdatePositionMarginService().
		Symbol(pp.symbol.Symbol).Type(typeMargin).
		Amount(utils.ConvFloat64ToStrDefault(float64(amountMargin))).Do(context.Background())
}

func (pp *PairProcessor) ClosePosition(risk *futures.PositionRisk) (err error) {
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		_, err = pp.CreateOrder(futures.OrderTypeTakeProfitMarket, futures.SideTypeBuy, futures.TimeInForceTypeGTC, 0, true, false, 0, 0, 0, 0)
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		_, err = pp.CreateOrder(futures.OrderTypeTakeProfitMarket, futures.SideTypeSell, futures.TimeInForceTypeGTC, 0, true, false, 0, 0, 0, 0)
	}
	return
}

func (pp *PairProcessor) GetPositionAmt() (positionAmt float64) {
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	positionAmt = utils.ConvStrToFloat64(risk.PositionAmt) // Convert string to float64
	return
}

func (pp *PairProcessor) GetPredictableUPnL(risk *futures.PositionRisk, price types.PriceType) (unRealizedProfit types.PriceType) {
	if risk == nil || pp.leverage <= 0 {
		return 0
	}
	entryPrice := types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
	positionAmt := types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	if positionAmt == 0 { // No position
		return 0
	} else if positionAmt < 0 { // Short position
		unRealizedProfit = types.PriceType(float64(entryPrice-price) * float64(positionAmt) * float64(pp.leverage))
	} else if positionAmt > 0 { // Long position
		unRealizedProfit = types.PriceType(float64(price-entryPrice) * float64(positionAmt) * float64(pp.leverage))
	}
	return
}
func (pp *PairProcessor) CheckAddPosition(risk *futures.PositionRisk, price types.PriceType) bool {
	if risk == nil {
		return false
	}
	positionAmt := types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt))
	liquidationPrice := types.PriceType(utils.ConvStrToFloat64(risk.LiquidationPrice))
	if positionAmt == 0 { // No position
		return true
	} else if positionAmt < 0 { // Short position
		return liquidationPrice > pp.GetUpBound() &&
			pp.GetPredictableUPnL(risk, pp.GetUpBound()) > -(pp.GetFreeBalance()*types.PriceType(pp.GetLeverage())) &&
			price <= pp.GetUpBound()
	} else if positionAmt > 0 { // Long position
		return liquidationPrice < pp.GetLowBound() &&
			pp.GetPredictableUPnL(risk, pp.GetLowBound()) > -(pp.GetFreeBalance()*types.PriceType(pp.GetLeverage())) &&
			price >= pp.GetLowBound()
	}
	return false
}

func (pp *PairProcessor) CheckStopLoss(free types.PriceType, risk *futures.PositionRisk, price types.PriceType) bool {
	if risk == nil || utils.ConvStrToFloat64(risk.PositionAmt) == 0 {
		return false
	}
	return (utils.ConvStrToFloat64(risk.PositionAmt) > 0 && price < pp.GetLowBound()) ||
		(utils.ConvStrToFloat64(risk.PositionAmt) < 0 && price > pp.GetUpBound()) ||
		types.PriceType(math.Abs(utils.ConvStrToFloat64(risk.UnRealizedProfit))) > free
}
