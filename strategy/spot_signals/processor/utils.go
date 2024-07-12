package processor

import (
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	utils "github.com/fr0ster/go-trading-utils/utils"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
)

func (pp *PairProcessor) RoundPrice(price float64) float64 {
	return utils.RoundToDecimalPlace(price, pp.GetTickSizeExp())
}

func (pp *PairProcessor) RoundQuantity(quantity float64) float64 {
	return utils.RoundToDecimalPlace(quantity, pp.GetStepSizeExp())
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
