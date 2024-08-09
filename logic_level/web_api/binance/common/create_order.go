package spot_web_api

import (
	"github.com/bitly/go-simplejson"
	order "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/order"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

func newPlaceOrder(apiKey, symbol, waHost, waPath string, sign signature.Sign) *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return order.New(apiKey, symbol, waHost, waPath, "order.place", sign)
}
