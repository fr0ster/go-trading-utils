package processor

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"

	"github.com/sirupsen/logrus"
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
func (pp *PairProcessor) createOrder(
	orderType futures.OrderType,
	sideType futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity items_types.QuantityType,
	closePosition bool,
	reduceOnly bool,
	price items_types.PriceType,
	stopPrice items_types.PriceType,
	activationPrice items_types.PriceType,
	callbackRate items_types.PricePercentType,
	times int,
	oldErr ...error) (
	order *futures.CreateOrderResponse, err error) {
	if times == 0 {
		if len(oldErr) == 0 {
			err = fmt.Errorf("can't create order")
		} else {
			err = oldErr[0]
		}
		return
	}
	pp.symbol, err = (*pp.pairInfo).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	if _, ok := pp.orderTypes[orderType]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.symbol.Symbol)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderResponseType(futures.NewOrderRespTypeRESULT).
			Symbol(string(futures.SymbolType(pp.symbol.Symbol))).
			Type(orderType).
			Side(sideType)
	if reduceOnly && !closePosition {
		service = service.ReduceOnly(reduceOnly)
	}
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == futures.OrderTypeMarket {
		// MARKET	quantity
		service = service.Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound))
	} else if orderType == futures.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound))
	} else if orderType == futures.OrderTypeStop || orderType == futures.OrderTypeTakeProfit {
		// STOP/TAKE_PROFIT	quantity, price, stopPrice
		service = service.
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			Price(utils.ConvFloat64ToStr(float64(price), priceRound)).
			StopPrice(utils.ConvFloat64ToStr(float64(stopPrice), priceRound))
	} else if orderType == futures.OrderTypeStopMarket || orderType == futures.OrderTypeTakeProfitMarket {
		// STOP_MARKET/TAKE_PROFIT_MARKET	stopPrice
		service = service.
			StopPrice(utils.ConvFloat64ToStr(float64(stopPrice), priceRound))
		if closePosition {
			service = service.ClosePosition(closePosition)
		}
	} else if orderType == futures.OrderTypeTrailingStopMarket {
		// TRAILING_STOP_MARKET	quantity,callbackRate
		service = service.
			TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(utils.ConvFloat64ToStr(float64(quantity), quantityRound)).
			CallbackRate(utils.ConvFloat64ToStr(float64(callbackRate), priceRound))
		if stopPrice != 0 {
			service = service.
				ActivationPrice(utils.ConvFloat64ToStr(float64(activationPrice), priceRound))
		}
	}
	order, err = service.Do(context.Background())
	if err != nil {
		logrus.Errorf("Can't create order: %v", err)
		apiError, _ := utils.ParseAPIError(err)
		if apiError == nil {
			return
		} else if apiError.Code == -2022 {
			// -2022 ReduceOnly Order is rejected.
			err = nil
			return
		} else if apiError.Code == -2027 {
			// -2027 Exceeded the maximum allowable position at current leverage.
			err = nil
			return
		} else if apiError.Code == -2028 {
			// -2028 Leverage is smaller than permitted: insufficient margin balance.
			err = nil
			return
		} else if apiError.Code == -1007 {
			time.Sleep(1 * time.Second)
			orders, err := pp.GetOpenOrders()
			if err != nil {
				return nil, err
			}
			for _, order := range orders {
				if order.Symbol == pp.symbol.Symbol &&
					order.Side == sideType &&
					order.Price == utils.ConvFloat64ToStr(float64(price), priceRound) {
					return &futures.CreateOrderResponse{
						Symbol:                  order.Symbol,
						OrderID:                 order.OrderID,
						ClientOrderID:           order.ClientOrderID,
						Price:                   order.Price,
						OrigQuantity:            order.OrigQuantity,
						ExecutedQuantity:        order.ExecutedQuantity,
						CumQuote:                order.CumQuote,
						ReduceOnly:              order.ReduceOnly,
						Status:                  order.Status,
						StopPrice:               order.StopPrice,
						TimeInForce:             order.TimeInForce,
						Type:                    order.Type,
						Side:                    order.Side,
						UpdateTime:              order.UpdateTime,
						WorkingType:             order.WorkingType,
						ActivatePrice:           order.ActivatePrice,
						PriceRate:               order.PriceRate,
						AvgPrice:                order.AvgPrice,
						PositionSide:            order.PositionSide,
						ClosePosition:           order.ClosePosition,
						PriceProtect:            order.PriceProtect,
						PriceMatch:              order.PriceMatch,
						SelfTradePreventionMode: order.SelfTradePreventionMode,
						GoodTillDate:            order.GoodTillDate,
						CumQty:                  order.CumQuantity,
						OrigType:                order.OrigType,
					}, nil
				}
			}
			// На наступних кодах помилок можна спробувати ще раз
		} else if apiError.Code == -1008 || apiError.Code == -5028 {
			time.Sleep(3 * time.Second)
			return pp.createOrder(
				orderType,
				sideType,
				timeInForce,
				quantity,
				closePosition,
				reduceOnly,
				price,
				stopPrice,
				activationPrice,
				callbackRate,
				times-1,
				err)
		}
		return
	}
	return
}
func (pp *PairProcessor) CreateOrder(
	orderType futures.OrderType,
	sideType futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity items_types.QuantityType,
	closePosition bool,
	reduceOnly bool,
	price items_types.PriceType,
	stopPrice items_types.PriceType,
	activationPrice items_types.PriceType,
	callbackRate items_types.PricePercentType) (
	order *futures.CreateOrderResponse, err error) {
	return pp.createOrder(
		orderType,
		sideType,
		timeInForce,
		quantity,
		closePosition,
		reduceOnly,
		price,
		stopPrice,
		activationPrice,
		callbackRate,
		repeatTimes)
}

func (pp *PairProcessor) GetOpenOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.symbol.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.symbol.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *futures.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.symbol.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *futures.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.symbol.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (err error) {
	return pp.client.NewCancelAllOpenOrdersService().Symbol(pp.symbol.Symbol).Do(context.Background())
}
