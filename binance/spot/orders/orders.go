package orders

import (
	"context"

	"github.com/adshao/go-binance/v2"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

//  1. Order with type STOP, parameter timeInForce can be sent ( default GTC).
//  2. Order with type TAKE_PROFIT, parameter timeInForce can be sent ( default GTC).
//  3. Condition orders will be triggered when:
//     a) If parameterpriceProtectis sent as true:
//     when price reaches the stopPrice ，the difference rate between "MARK_PRICE" and
//     "CONTRACT_PRICE" cannot be larger than the "triggerProtect" of the symbol
//     "triggerProtect" of a symbol can be got from GET /fapi/v1/exchangeInfo
//     b) STOP, STOP_MARKET:
//     BUY: latest price ("MARK_PRICE" or "CONTRACT_PRICE") >= stopPrice
//     SELL: latest price ("MARK_PRICE" or "CONTRACT_PRICE") <= stopPrice
//     c) TAKE_PROFIT, TAKE_PROFIT_MARKET:
//     BUY: latest price ("MARK_PRICE" or "CONTRACT_PRICE") <= stopPrice
//     SELL: latest price ("MARK_PRICE" or "CONTRACT_PRICE") >= stopPrice
//     d) TRAILING_STOP_MARKET:
//     BUY: the lowest price after order placed <= activationPrice,
//     and the latest price >= the lowest price * (1 + callbackRate)
//     SELL: the highest price after order placed >= activationPrice,
//     and the latest price <= the highest price * (1 - callbackRate)
//  4. For TRAILING_STOP_MARKET, if you got such error code.
//     {"code": -2021, "msg": "Order would immediately trigger."}
//     means that the parameters you send do not meet the following requirements:
//     BUY: activationPrice should be smaller than latest price.
//     SELL: activationPrice should be larger than latest price.
//     If newOrderRespType is sent as RESULT :
//     MARKET order: the final FILLED result of the order will be return directly.
//     LIMIT order with special timeInForce:
//     the final status result of the order(FILLED or EXPIRED)
//     will be returned directly.
//  5. STOP_MARKET, TAKE_PROFIT_MARKET with closePosition=true:
//     Follow the same rules for condition orders.
//     If triggered，close all current long position( if SELL) or current short position( if BUY).
//     Cannot be used with quantity parameter
//     Cannot be used with reduceOnly parameter
//     In Hedge Mode,cannot be used with BUY orders in LONG position side
//     and cannot be used with SELL orders in SHORT position side
//  6. selfTradePreventionMode is only effective when timeInForce set to IOC or GTC or GTD.
//  7. In extreme market conditions,
//     timeInForce GTD order auto cancel time might be delayed comparing to goodTillDate
func createOrder(
	client *binance.Client, // 1
	symbol string, // 2
	quantityRound int, // 3
	priceRound int, // 4
	orderType binance.OrderType, // 5
	sideType binance.SideType, // 6
	timeInForce binance.TimeInForceType, // 7
	quantity items_types.QuantityType, // 8
	// closePosition bool, // 9
	// reduceOnly bool, // 10
	price items_types.PriceType, // 11
	stopPrice items_types.PriceType, // 12
	// activationPrice items_types.PriceType, // 13
	callbackRate items_types.PricePercentType) ( // 14
	order *binance.CreateOrderResponse, err error) {
	service :=
		client.NewCreateOrderService().
			NewOrderRespType(binance.NewOrderRespTypeRESULT).
			Symbol(string(binance.SymbolType(symbol))).
			Type(orderType).
			Side(sideType)
	// if reduceOnly && !closePosition {
	// 	service = service.ReduceOnly(reduceOnly)
	// }
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == binance.OrderTypeMarket {
		// MARKET	quantity
		service = service.Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound))
	} else if orderType == binance.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound))
	} else if orderType == binance.OrderTypeStopLossLimit || orderType == binance.OrderTypeTakeProfitLimit {
		// STOP/TAKE_PROFIT	quantity, price, stopPrice
		service = service.
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound)).
			StopPrice(utils.ConvFloat64ToStr(float64(stopPrice), priceRound)).
			TrailingDelta(utils.ConvFloat64ToStr(float64(callbackRate), priceRound))
	} else if orderType == binance.OrderTypeStopLoss || orderType == binance.OrderTypeTakeProfit {
		// STOP_MARKET/TAKE_PROFIT_MARKET	stopPrice
		service = service.
			StopPrice(utils.ConvFloat64ToStr(float64(stopPrice), priceRound))
		// if closePosition {
		// 	service = service.ClosePosition(closePosition)
		// }
		// } else if orderType == binance.OrderTypeTrailingStopMarket {
		// 	// TRAILING_STOP_MARKET	quantity,callbackRate
		// 	service = service.
		// 		TimeInForce(binance.TimeInForceTypeGTC).
		// 		Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
		// 		CallbackRate(utils.ConvFloat64ToStr(float64(callbackRate), priceRound))
		// 	if stopPrice != 0 {
		// 		service = service.
		// 			ActivationPrice(utils.ConvFloat64ToStr(float64(activationPrice), priceRound))
		// 	}
	}
	order, err = service.Do(context.Background())
	return
}

func CreateOrderCreator(
	client *binance.Client,
	symbol string,
	quantityRound int,
	priceRound int) func(*orders_types.Orders) orders_types.CreateOrderFunction {
	return func(*orders_types.Orders) orders_types.CreateOrderFunction {
		return func(
			orderType orders_types.OrderType,
			sideType orders_types.SideType,
			timeInForce orders_types.TimeInForceType,
			quantity items_types.QuantityType,
			closePosition bool,
			reduceOnly bool,
			price items_types.PriceType,
			stopPrice items_types.PriceType,
			activationPrice items_types.PriceType,
			callbackRate items_types.PricePercentType) (response orders_types.CreateOrderResponse, err error) {
			var orders *binance.CreateOrderResponse
			orders, err = createOrder(
				client,
				symbol,
				quantityRound,
				priceRound,
				binance.OrderType(orderType),
				binance.SideType(sideType),
				binance.TimeInForceType(timeInForce),
				quantity,
				// closePosition,
				// reduceOnly,
				price,
				stopPrice,
				// activationPrice,
				callbackRate)
			response = orders_types.CreateOrderResponse{
				Symbol:           orders.Symbol,
				OrderID:          orders.OrderID,
				ClientOrderID:    orders.ClientOrderID,
				Price:            orders.Price,
				OrigQuantity:     orders.OrigQuantity,
				ExecutedQuantity: orders.ExecutedQuantity,
				Status:           orders_types.OrderStatusType(orders.Status),
				// StopPrice:        orders.StopPrice,
				TimeInForce: orders_types.TimeInForceType(orders.TimeInForce),
				Type:        orders_types.OrderType(orders.Type),
				Side:        orders_types.SideType(orders.Side),
				// UpdateTime:       orders.UpdateTime,
			}
			return response, err
		}
	}
}

func futures2orders(input *binance.Order) *orders_types.Order {
	return &orders_types.Order{
		Symbol:        input.Symbol,
		OrderID:       input.OrderID,
		ClientOrderID: input.ClientOrderID,
		Price:         input.Price,
		// ReduceOnly:              input.ReduceOnly,
		OrigQuantity:     input.OrigQuantity,
		ExecutedQuantity: input.ExecutedQuantity,
		CumQuantity:      input.CummulativeQuoteQuantity,
		CumQuote:         input.CummulativeQuoteQuantity,
		Status:           orders_types.OrderStatusType(input.Status),
		TimeInForce:      orders_types.TimeInForceType(input.TimeInForce),
		Type:             orders_types.OrderType(input.Type),
		Side:             orders_types.SideType(input.Side),
		StopPrice:        input.StopPrice,
		Time:             input.Time,
		UpdateTime:       input.UpdateTime,
		// WorkingType:             orders_types.WorkingType(input.WorkingType),
		// ActivatePrice:           input.ActivatePrice,
		// PriceRate:               input.PriceRate,
		// AvgPrice:                input.AvgPrice,
		// OrigType:                orders_types.OrderType(input.OrigType),
		// PositionSide:            orders_types.PositionSideType(input.PositionSide),
		// PriceProtect:            input.PriceProtect,
		// ClosePosition:           input.ClosePosition,
		// PriceMatch:              input.PriceMatch,
		// SelfTradePreventionMode: input.SelfTradePreventionMode,
		// GoodTillDate:            input.GoodTillDate,
	}
}

func GetOpenOrdersCreator(client *binance.Client) func(pp *orders_types.Orders) func() ([]*orders_types.Order, error) {
	return func(orders *orders_types.Orders) func() ([]*orders_types.Order, error) {
		return func() ([]*orders_types.Order, error) {
			var arrOrders []*orders_types.Order
			futuresOrders, err := client.NewListOpenOrdersService().Symbol(orders.Symbol()).Do(context.Background())
			if err != nil {
				return nil, err
			}
			for _, order := range futuresOrders {
				arrOrders = append(arrOrders, futures2orders(order))
			}
			return arrOrders, err
		}
	}
}

func GetAllOrdersCreator(client *binance.Client) func(pp *orders_types.Orders) func() ([]*orders_types.Order, error) {
	return func(orders *orders_types.Orders) func() (orders []*orders_types.Order, err error) {
		return func() ([]*orders_types.Order, error) {
			var arrOrders []*orders_types.Order
			futuresOrders, err := client.NewListOrdersService().Symbol(orders.Symbol()).Do(context.Background())
			if err != nil {
				return nil, err
			}
			for _, order := range futuresOrders {
				arrOrders = append(arrOrders, futures2orders(order))
			}
			return arrOrders, err
		}
	}
}

func GetOrderCreator(client *binance.Client) func(pp *orders_types.Orders) func(orderID int64) (*orders_types.Order, error) {
	return func(orders *orders_types.Orders) func(orderID int64) (*orders_types.Order, error) {
		return func(orderID int64) (*orders_types.Order, error) {
			futuresOrder, err := client.NewGetOrderService().Symbol(orders.Symbol()).OrderID(orderID).Do(context.Background())
			if err != nil {
				return nil, err
			}
			return futures2orders(futuresOrder), nil
		}
	}
}

func CancelOrderCreator(client *binance.Client) func(pp *orders_types.Orders) func(orderID int64) (*orders_types.CancelOrderResponse, error) {
	return func(orders *orders_types.Orders) func(orderID int64) (*orders_types.CancelOrderResponse, error) {
		return func(orderID int64) (*orders_types.CancelOrderResponse, error) {
			response, err := client.NewCancelOrderService().Symbol(orders.Symbol()).OrderID(orderID).Do(context.Background())
			if err != nil {
				return nil, err
			}
			return &orders_types.CancelOrderResponse{
				ClientOrderID:    response.ClientOrderID,
				CumQuantity:      response.CummulativeQuoteQuantity,
				CumQuote:         response.CummulativeQuoteQuantity,
				ExecutedQuantity: response.ExecutedQuantity,
				OrderID:          response.OrderID,
				OrigQuantity:     response.OrigQuantity,
				Price:            response.Price,
				// ReduceOnly:       response.ReduceOnly,
				Side:   orders_types.SideType(response.Side),
				Status: orders_types.OrderStatusType(response.Status),
				// StopPrice:        response.StopPrice,
				Symbol:      response.Symbol,
				TimeInForce: orders_types.TimeInForceType(response.TimeInForce),
				Type:        orders_types.OrderType(response.Type),
				// UpdateTime:       response.UpdateTime,
				// WorkingType:      orders_types.WorkingType(response.WorkingType),
				// ActivatePrice:    response.ActivatePrice,
				// PriceRate:        response.PriceRate,
				// OrigType:         response.OrigType,
				// PositionSide:     orders_types.PositionSideType(response.PositionSide),
				// PriceProtect:     response.PriceProtect,
			}, nil
		}
	}
}

func CancelAllOrdersCreator(client *binance.Client) func(pp *orders_types.Orders) func() error {
	return func(orders *orders_types.Orders) func() error {
		return func() (err error) {
			_, err = client.NewCancelOpenOrdersService().Symbol(orders.Symbol()).Do(context.Background())
			return
		}
	}
}
