package processor

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"

	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	exchange "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) RoundPrice(price items.PriceType) items.PriceType {
	return items.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *PairProcessor) RoundQuantity(quantity items.QuantityType) items.QuantityType {
	return items.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
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

func (pp *PairProcessor) GetNextUpCoefficient() items.PriceType {
	if pp.depth.GetAsks() == nil ||
		pp.depth.GetBids() == nil ||
		pp.depth.GetAsks().GetSummaValue() == 0 ||
		pp.depth.GetBids().GetSummaValue() == 0 {
		return 1
	}
	coefficients := items.PriceType(pp.depth.GetAsks().GetSummaValue() / pp.depth.GetBids().GetSummaValue())
	if coefficients > 1 {
		return 1
	} else {
		return coefficients
	}
}

func (pp *PairProcessor) GetNextDownCoefficient() items.PriceType {
	if pp.depth.GetAsks() == nil ||
		pp.depth.GetBids() == nil ||
		pp.depth.GetAsks().GetSummaValue() == 0 ||
		pp.depth.GetBids().GetSummaValue() == 0 {
		return 1
	}
	coefficients := items.PriceType(pp.depth.GetAsks().GetSummaValue() / pp.depth.GetBids().GetSummaValue())
	if coefficients > 1 {
		return coefficients
	} else {
		return 1
	}
}

func (pp *PairProcessor) GetTargetPrices(price ...items.PriceType) (priceUp, priceDown items.PriceType, err error) {
	var currentPrice items.PriceType
	if len(price) == 0 {
		currentPrice, err = pp.GetCurrentPrice()
		if err != nil {
			return
		}
	} else {
		currentPrice = price[0]
	}
	priceUp = pp.RoundPrice(currentPrice * (1 + pp.GetDeltaPrice()*pp.GetNextUpCoefficient()))
	priceDown = pp.RoundPrice(currentPrice * (1 - pp.GetDeltaPrice()*pp.GetNextDownCoefficient()))
	return
}

func (pp *PairProcessor) GetLimitPrices() (priceUp, priceDown items.PriceType) {
	var (
		askMax *items.DepthItem
		bidMax *items.DepthItem
	)
	_, askMax = pp.depth.GetAsks().GetMinMaxByQuantity()
	_, bidMax = pp.depth.GetBids().GetMinMaxByQuantity()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	return
}

func (pp *PairProcessor) GetPrices(
	price items.PriceType,
	isDynamic bool) (
	priceUp items.PriceType,
	quantityUp items.QuantityType,
	priceDown items.PriceType,
	quantityDown items.QuantityType,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	if pp.depth != nil {
		priceUp, priceDown, err = pp.GetTargetPrices()
	} else {
		priceUp = items.PriceType(pp.RoundPrice(price * (1 + pp.GetDeltaPrice())))
		priceDown = items.PriceType(pp.RoundPrice(price * (1 - pp.GetDeltaPrice())))
	}
	reduceOnlyUp = false
	reduceOnlyDown = false

	quantityUp = items.QuantityType(pp.RoundQuantity(items.QuantityType(pp.GetLimitOnTransaction() / priceUp)))
	quantityDown = items.QuantityType(pp.RoundQuantity(items.QuantityType(pp.GetLimitOnTransaction() / priceDown)))

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

func LimitRead(degree int, symbols []string, client *binance.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange.RateLimits,
	dayOrderLimit *exchange.RateLimits,
	minuteRawRequestLimit *exchange.RateLimits) {
	exchangeInfo := exchange.New()
	spot_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
