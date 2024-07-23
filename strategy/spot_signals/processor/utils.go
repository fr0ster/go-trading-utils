package processor

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) RoundValue(value items_types.ValueType) items_types.ValueType {
	return items_types.ValueType(utils.RoundToDecimalPlace(float64(value), pp.GetTickSizeExp()))
}

func (pp *PairProcessor) RoundPrice(price items_types.PriceType) items_types.PriceType {
	return items_types.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *PairProcessor) RoundQuantity(quantity items_types.QuantityType) items_types.QuantityType {
	return items_types.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
}

func (pp *PairProcessor) Debug(fl, id string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		orders, _ := pp.GetOpenOrders()
		logrus.Debugf("%s %s %s:", fl, id, pp.symbol.Symbol)
		for _, order := range orders {
			logrus.Debugf(" Open Order %v on price %v OrderSide %v Status %s", order.OrderID, order.Price, order.Side, order.Status)
		}
	}
}

func (pp *PairProcessor) GetNextUpCoefficient() items_types.PriceType {
	if pp.depth.GetAsks() == nil ||
		pp.depth.GetBids() == nil ||
		pp.depth.GetAsks().GetSummaValue() == 0 ||
		pp.depth.GetBids().GetSummaValue() == 0 {
		return 1
	}
	coefficients := items_types.PriceType(pp.depth.GetAsks().GetSummaValue() / pp.depth.GetBids().GetSummaValue())
	if coefficients > 1 {
		return 1
	} else {
		return coefficients
	}
}

func (pp *PairProcessor) GetNextDownCoefficient() items_types.PriceType {
	if pp.depth.GetAsks() == nil ||
		pp.depth.GetBids() == nil ||
		pp.depth.GetAsks().GetSummaValue() == 0 ||
		pp.depth.GetBids().GetSummaValue() == 0 {
		return 1
	}
	coefficients := items_types.PriceType(pp.depth.GetAsks().GetSummaValue() / pp.depth.GetBids().GetSummaValue())
	if coefficients > 1 {
		return coefficients
	} else {
		return 1
	}
}

func (pp *PairProcessor) GetTargetPrices(price ...items_types.PriceType) (priceDown, priceUp items_types.PriceType, err error) {
	priceUp = pp.NextPriceUp(price...)
	priceDown = pp.NextPriceDown(price...)
	return
}

func (pp *PairProcessor) GetLimitPrices(price ...items_types.PriceType) (priceTargetDown, priceTargetUp, priceDown, priceUp items_types.PriceType, err error) {
	var (
		askMax *items_types.DepthItem
		bidMax *items_types.DepthItem
	)
	priceTargetDown, priceTargetUp, err = pp.GetTargetPrices(price...)
	if err != nil {
		return
	}
	asksFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() > priceTargetUp
	}
	bidsFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() < priceTargetDown
	}
	_, askMax = pp.depth.GetAsks().GetFiltered(asksFilter).GetMinMaxByValue()
	_, bidMax = pp.depth.GetBids().GetFiltered(bidsFilter).GetMinMaxByValue()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	return
}

func (pp *PairProcessor) GetPrices(
	price items_types.PriceType,
	isDynamic bool) (
	priceUp items_types.PriceType,
	quantityUp items_types.QuantityType,
	priceDown items_types.PriceType,
	quantityDown items_types.QuantityType,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	priceDown, priceUp, err = pp.GetTargetPrices(price)
	if err != nil {
		return
	}
	reduceOnlyUp = false
	reduceOnlyDown = false

	quantityUp = items_types.QuantityType(pp.RoundQuantity(items_types.QuantityType(pp.GetLimitOnTransaction() / priceUp)))
	quantityDown = items_types.QuantityType(pp.RoundQuantity(items_types.QuantityType(pp.GetLimitOnTransaction() / priceDown)))

	if quantityUp == 0 && quantityDown == 0 {
		err = fmt.Errorf("can't calculate initial position for price up %v and price down %v", priceUp, priceDown)
		return
	}
	if float64(quantityUp)*float64(priceUp) < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity up %v * price up %v < notional %v", quantityUp, priceUp, pp.GetNotional())
		return
	} else if float64(quantityDown)*float64(priceDown) < pp.GetNotional() {
		err = fmt.Errorf("calculated quantity down %v * price down %v < notional %v", quantityDown, priceDown, pp.GetNotional())
		return
	}
	return
}

func (pp *PairProcessor) LimitRead(degree int, symbols []string, client *binance.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange.RateLimits,
	dayOrderLimit *exchange.RateLimits,
	minuteRawRequestLimit *exchange.RateLimits) {
	minuteOrderLimit = pp.exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = pp.exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = pp.exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
