package processor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
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
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	pp.leverage = int(utils.ConvStrToFloat64(risk.Leverage))
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

func (pp *PairProcessor) SetPositionMargin(amountMargin float64, typeMargin int) (err error) {
	return pp.client.NewUpdatePositionMarginService().
		Symbol(pp.symbol.Symbol).Type(typeMargin).
		Amount(utils.ConvFloat64ToStrDefault(amountMargin)).Do(context.Background())
}

func (pp *PairProcessor) ClosePosition(risk *futures.PositionRisk) (err error) {
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		_, err = pp.CreateOrder(futures.OrderTypeTakeProfitMarket, futures.SideTypeBuy, futures.TimeInForceTypeGTC, 0, true, false, 0, 0, 0, 0)
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		_, err = pp.CreateOrder(futures.OrderTypeTakeProfitMarket, futures.SideTypeSell, futures.TimeInForceTypeGTC, 0, true, false, 0, 0, 0, 0)
	}
	return
}
