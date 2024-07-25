package processor

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"

	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/depths"
	spot_exchangeinfo "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_orders "github.com/fr0ster/go-trading-utils/binance/spot/orders"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchangeinfo_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	processor_types "github.com/fr0ster/go-trading-utils/types/processor"
)

func New(
	client *binance.Client,
	degree int,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpAndLowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	callbackRate items_types.PricePercentType,
	depthAPILimit depth_types.DepthAPILimit,
	ordersCallBack func(p *processor_types.Processor) func(o *orders_types.Orders) binance.WsUserDataHandler,
	ordersErrHandler func(p *processor_types.Processor) func(o *orders_types.Orders) binance.ErrHandler,
	depthsCallBack func(p *processor_types.Processor) func(d *depth_types.Depths) binance.WsDepthHandler,
	depthsErrHandler func(p *processor_types.Processor) func(d *depth_types.Depths) binance.ErrHandler,
	debug bool,
	quits ...chan struct{},
) (pairProcessor *processor_types.Processor, err error) {
	var quit chan struct{}
	if len(quits) > 0 {
		quit = quits[0]
	} else {
		quit = make(chan struct{})
	}
	exchange := exchangeinfo_types.New(spot_exchangeinfo.InitCreator(client, degree, symbol))
	symbolInfo := exchange.GetSymbols().GetSymbol(symbol)
	pairProcessor, err = processor_types.New(
		quit,     // quit
		symbol,   // pair
		exchange, // exchange
		func(p *processor_types.Processor) processor_types.DepthConstructor {
			return func() *depth_types.Depths {
				var (
					callBack   func(d *depth_types.Depths) binance.WsDepthHandler
					errHandler func(*depth_types.Depths) binance.ErrHandler
				)
				if depthsCallBack != nil {
					callBack = spot_depth.CallBackCreator(depthsCallBack(p))
				} else {
					callBack = spot_depth.CallBackCreator()
				}
				if depthsErrHandler != nil {
					errHandler = spot_depth.WsErrorHandlerCreator(depthsErrHandler(p))
				} else {
					errHandler = spot_depth.WsErrorHandlerCreator()
				}
				return depth_types.New(
					degree,
					symbol,
					spot_depth.DepthStreamCreator(
						callBack,
						errHandler),
					spot_depth.InitCreator(depthAPILimit, client))
			}
		}, // depthsCreator
		func(p *processor_types.Processor) processor_types.OrdersConstructor {
			return func() *orders_types.Orders {
				var (
					callBack   func(*orders_types.Orders) binance.WsUserDataHandler
					errHandler func(*orders_types.Orders) binance.ErrHandler
				)
				if depthsCallBack != nil {
					callBack = spot_orders.CallBackCreator(ordersCallBack(p))
				} else {
					callBack = spot_orders.CallBackCreator()
				}
				if depthsErrHandler != nil {
					errHandler = spot_orders.WsErrorHandlerCreator(ordersErrHandler(p))
				} else {
					errHandler = spot_orders.WsErrorHandlerCreator()
				}
				return orders_types.New(
					symbol, // symbol
					spot_orders.UserDataStreamCreator(
						client,
						callBack,
						errHandler), // userDataStream
					spot_orders.CreateOrderCreator(
						client,
						int(float64(symbolInfo.GetStepSize())),
						int(float64(symbolInfo.GetTickSize()))), // createOrder
					spot_orders.GetOpenOrdersCreator(client),   // getOpenOrders
					spot_orders.GetAllOrdersCreator(client),    // getAllOrders
					spot_orders.GetOrderCreator(client),        // getOrder
					spot_orders.CancelOrderCreator(client),     // cancelOrder
					spot_orders.CancelAllOrdersCreator(client), // cancelAllOrders
					quit)
			}
		}, // orders
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Balances {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.Free) + utils.ConvStrToFloat64(asset.Locked))
				}
			}
			return 0.0
		}, // getBaseBalance
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Balances {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetTargetSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.Free) + utils.ConvStrToFloat64(asset.Locked))
				}
			}
			return 0.0
		}, // getTargetBalance
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Balances {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.Free))
				}
			}
			return 0.0
		}, // getFreeBalance
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Balances {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.Locked))
				}
			}
			return 0.0
		}, // getLockedBalance
		func() items_types.PriceType {
			price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
			if err != nil {
				return 0
			}
			return items_types.PriceType(utils.ConvStrToFloat64(price[0].Price))
		}, // getCurrentPrice
		nil, // getPositionRisk
		nil, // getLeverage
		nil, // setLeverage
		nil, // getMarginType
		nil, // setMarginType
		nil, // setPositionMargin
		nil, // closePosition
		func() items_types.PricePercentType {
			return deltaPrice
		}, // getDeltaPrice
		func() items_types.QuantityPercentType {
			return deltaQuantity
		}, // getDeltaQuantity
		func() items_types.ValueType {
			return limitOnPosition
		}, // getLimitOnPosition
		func() items_types.ValuePercentType {
			return limitOnTransaction
		}, // getLimitOnTransaction
		func() items_types.PricePercentType {
			return UpAndLowBound
		}, // getUpAndLowBound
		func() items_types.PricePercentType {
			return callbackRate
		}, // getCallbackRate
		debug)
	if err != nil {
		logrus.Errorf("Can't init pair: %v", err)
		close(quit)
		return
	}
	return
}
