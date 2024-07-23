package processor

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

//  1. LIMIT_MAKER are LIMIT orders that will be rejected if they would immediately match and trade as a taker.
//  2. STOP_LOSS and TAKE_PROFIT will execute a MARKET order when the stopPrice is reached.
//     Any LIMIT or LIMIT_MAKER type order can be made an iceberg order by sending an icebergQty.
//     Any order with an icebergQty MUST have timeInForce set to GTC.
//  3. MARKET orders using the quantity field specifies the amount of the base asset the user wants to buy or sell at the market price.
//     For example, sending a MARKET order on BTCUSDT will specify how much BTC the user is buying or selling.
//  4. MARKET orders using quoteOrderQty specifies the amount the user wants to spend (when buying) or receive (when selling) the quote asset;
//     the correct quantity will be determined based on the market liquidity and quoteOrderQty.
//     Using BTCUSDT as an example:
//     On the BUY side, the order will buy as many BTC as quoteOrderQty USDT can.
//     On the SELL side, the order will sell as much BTC needed to receive quoteOrderQty USDT.
//  5. MARKET orders using quoteOrderQty will not break LOT_SIZE filter rules; the order will execute a quantity that will have the notional value as close as possible to quoteOrderQty.
//     same newClientOrderId can be accepted only when the previous one is filled, otherwise the order will be rejected.
//  6. For STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT and TAKE_PROFIT orders, trailingDelta can be combined with stopPrice.
//
//  7. Trigger order price rules against market price for both MARKET and LIMIT versions:
//     Price above market price: STOP_LOSS BUY, TAKE_PROFIT SELL
//     Price below market price: STOP_LOSS SELL, TAKE_PROFIT BUY
func (pp *PairProcessor) createOrder(
	orderType binance.OrderType, // MARKET, LIMIT, LIMIT_MAKER, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	sideType binance.SideType, // BUY, SELL
	timeInForce binance.TimeInForceType, // GTC, IOC, FOK
	quantity items_types.QuantityType, // BTC for example if we buy or sell BTC
	quantityQty items_types.PriceType, // USDT for example if we buy or sell BTC
	// price for 1 BTC
	// it's price of order execution for LIMIT, LIMIT_MAKER
	// after execution of STOP_LOSS, TAKE_PROFIT, wil be created MARKET order
	// after execution of STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT wil be created LIMIT order with price of order execution from PRICE parameter
	price items_types.PriceType,
	// price for stop loss or take profit it's price of order execution for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	stopPrice items_types.PriceType,
	// trailingDelta for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	// https://github.com/binance/binance-spot-api-docs/blob/master/faqs/trailing-stop-faq.md
	trailingDelta int,
	times int) (
	order *binance.CreateOrderResponse, err error) {
	if times == 0 {
		err = fmt.Errorf("can't create order")
		return
	}
	if _, ok := pp.orderTypes[orderType]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.pairInfo.Symbol)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / float64(pp.pairInfo.GetStepSize())))
		priceRound    = int(math.Log10(1 / float64(pp.pairInfo.GetTickSizeExp())))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderRespType(binance.NewOrderRespTypeRESULT).
			Symbol(string(binance.SymbolType(pp.pairInfo.Symbol))).
			Type(orderType).
			Side(sideType)
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == binance.OrderTypeMarket {
		// MARKET	quantity or quoteOrderQty
		if quantity != 0 {
			service = service.
				Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound))
		} else if quantityQty != 0 {
			service = service.
				QuoteOrderQty(utils.ConvFloat64ToStr(float64(quantityQty), quantityRound))
		} else {
			err = fmt.Errorf("quantity or quoteOrderQty must be set")
			return
		}
	} else if orderType == binance.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound))
	} else if orderType == binance.OrderTypeLimitMaker {
		// LIMIT_MAKER	quantity, price
		service = service.
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound))
	} else if orderType == binance.OrderTypeStopLoss || orderType == binance.OrderTypeTakeProfit {
		// STOP_LOSS/TAKE_PROFIT quantity, stopPrice or trailingDelta
		service = service.
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound))
		if stopPrice != 0 {
			service = service.StopPrice(utils.ConvFloat64ToStr(float64(price), priceRound))
		} else if trailingDelta != 0 {
			service = service.TrailingDelta(strconv.Itoa(trailingDelta))
		} else {
			err = fmt.Errorf("stopPrice or trailingDelta must be set")
			return
		}
	} else if orderType == binance.OrderTypeStopLossLimit || orderType == binance.OrderTypeTakeProfitLimit {
		// STOP_LOSS_LIMIT/TAKE_PROFIT_LIMIT timeInForce, quantity, price, stopPrice or trailingDelta
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound))
		if stopPrice != 0 {
			service = service.StopPrice(utils.ConvFloat64ToStr(float64(price), priceRound))
		} else if trailingDelta != 0 {
			service = service.TrailingDelta(strconv.Itoa(trailingDelta))
		} else {
			err = fmt.Errorf("stopPrice or trailingDelta must be set")
			return
		}
	}
	order, err = service.Do(context.Background())
	if err != nil {
		apiError, _ := utils.ParseAPIError(err)
		if apiError == nil {
			return
		}
		if apiError.Code == -1007 {
			time.Sleep(1 * time.Second)
			orders, err := pp.GetOpenOrders()
			if err != nil {
				return nil, err
			}
			for _, order := range orders {
				if order.Symbol == pp.pairInfo.Symbol && order.Side == sideType && order.Price == utils.ConvFloat64ToStr(float64(price), priceRound) {
					return &binance.CreateOrderResponse{
						Symbol:                   order.Symbol,
						OrderID:                  order.OrderID,
						ClientOrderID:            order.ClientOrderID,
						Price:                    order.Price,
						OrigQuantity:             order.OrigQuantity,
						ExecutedQuantity:         order.ExecutedQuantity,
						CummulativeQuoteQuantity: order.CummulativeQuoteQuantity,
						IsIsolated:               order.IsIsolated,
						Status:                   order.Status,
						TimeInForce:              order.TimeInForce,
						Type:                     order.Type,
						Side:                     order.Side,
					}, nil
				}
			}
		} else if apiError.Code == -1008 {
			time.Sleep(3 * time.Second)
			return pp.createOrder(orderType, sideType, timeInForce, quantity, quantityQty, price, stopPrice, trailingDelta, times-1)
		}
	}
	return
}

func (pp *PairProcessor) CreateOrder(
	orderType binance.OrderType, // MARKET, LIMIT, LIMIT_MAKER, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	sideType binance.SideType, // BUY, SELL
	timeInForce binance.TimeInForceType, // GTC, IOC, FOK
	quantity items_types.QuantityType, // BTC for example if we buy or sell BTC
	quantityQty items_types.PriceType, // USDT for example if we buy or sell BTC
	// price for 1 BTC
	// it's price of order execution for LIMIT, LIMIT_MAKER
	// after execution of STOP_LOSS, TAKE_PROFIT, wil be created MARKET order
	// after execution of STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT wil be created LIMIT order with price of order execution from PRICE parameter
	price items_types.PriceType,
	// price for stop loss or take profit it's price of order execution for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	stopPrice items_types.PriceType,
	// trailingDelta for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	// https://github.com/binance/binance-spot-api-docs/blob/master/faqs/trailing-stop-faq.md
	trailingDelta int) (
	order *binance.CreateOrderResponse, err error) {
	return pp.createOrder(orderType, sideType, timeInForce, quantity, quantityQty, price, stopPrice, trailingDelta, 3)
}

func (pp *PairProcessor) CheckOrderType(orderType binance.OrderType) bool {
	_, ok := pp.orderTypes[orderType]
	return ok
}

func (pp *PairProcessor) GetOpenOrders() (orders []*binance.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*binance.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *binance.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.pairInfo.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *binance.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.pairInfo.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (orders *binance.CancelOpenOrdersResponse, err error) {
	return pp.client.NewCancelOpenOrdersService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
}
