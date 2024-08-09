package spot_web_api

import (
	"github.com/bitly/go-simplejson"
	"github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/order"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Функція для створення нової структури CancelReplaceOrderParams
func newCancelReplaceOrder(apiKey, symbol string, waHost web_api.WsHost, waPath web_api.WsPath, sign signature.Sign) *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return order.New(apiKey, symbol, "order.cancelReplace", waHost, waPath, sign)
}
