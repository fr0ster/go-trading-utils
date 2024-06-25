package futures_signals

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
)

func LimitRead(degree int, symbols []string, client *futures.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	err error) {
	exchangeInfo := exchange_types.New()
	futures_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	if minuteRawRequestLimit == nil {
		err = fmt.Errorf("minute raw request limit is not found")
		return
	}
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}
