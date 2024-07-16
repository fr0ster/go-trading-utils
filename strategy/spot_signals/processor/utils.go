package processor

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	utils "github.com/fr0ster/go-trading-utils/utils"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
)

func (pp *PairProcessor) RoundPrice(price types.PriceType) types.PriceType {
	return types.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *PairProcessor) RoundQuantity(quantity types.QuantityType) types.QuantityType {
	return types.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
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

func (pp *PairProcessor) GetTargetPrices() (priceUp, priceDown types.PriceType, err error) {
	if pp.depth != nil {
		priceUp, priceDown, _, _ = pp.depth.GetTargetPrices(pp.depth.GetPercentToTarget())
	} else {
		err = fmt.Errorf("depth is nil")
	}
	return
}

func (pp *PairProcessor) GetLimitPrices() (priceUp, priceDown types.PriceType, err error) {
	var (
		askMax *types.DepthItem
		bidMax *types.DepthItem
	)
	if pp.depth != nil {
		askMax, err = pp.depth.AskMax()
		if err != nil {
			return
		}
		bidMax, err = pp.depth.BidMax()
		if err != nil {
			return
		}
		priceUp = askMax.GetPrice()
		priceDown = bidMax.GetPrice()
	} else {
		err = fmt.Errorf("depth is nil")
	}
	return
}

func (pp *PairProcessor) GetPrices(
	price types.PriceType,
	isDynamic bool) (
	priceUp types.PriceType,
	quantityUp types.QuantityType,
	priceDown types.PriceType,
	quantityDown types.QuantityType,
	reduceOnlyUp bool,
	reduceOnlyDown bool,
	err error) {
	if pp.depth != nil {
		priceUp, priceDown, err = pp.GetTargetPrices()
	} else {
		priceUp = types.PriceType(pp.RoundPrice(price * (1 + pp.GetDeltaPrice())))
		priceDown = types.PriceType(pp.RoundPrice(price * (1 - pp.GetDeltaPrice())))
	}
	reduceOnlyUp = false
	reduceOnlyDown = false

	quantityUp = types.QuantityType(pp.RoundQuantity(types.QuantityType(pp.GetLimitOnTransaction() / priceUp)))
	quantityDown = types.QuantityType(pp.RoundQuantity(types.QuantityType(pp.GetLimitOnTransaction() / priceDown)))

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
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	exchangeInfo := exchange_types.New()
	spot_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
