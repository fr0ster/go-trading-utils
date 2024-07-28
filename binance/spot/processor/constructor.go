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
	baseSymbol := string(exchange.GetSymbol(symbol).GetBaseSymbol())
	targetSymbol := string(exchange.GetSymbol(symbol).GetTargetSymbol())
	pairProcessor, err = processor_types.New(
		quit,     // quit
		symbol,   // pair
		exchange, // exchange
		depthsCreator(
			client, // client
			degree, // degree
			depthAPILimit,
			depthsCallBack,   // depthsCallBack
			depthsErrHandler, // depthsCreator
		), // depthsCreator
		ordersCreator(
			client,           // client
			ordersCallBack,   // ordersCallBack
			ordersErrHandler, // ordersErrHandler
			quit,             // ordersCreator
		), // orders
		getBaseBalance(
			client, // client
			baseSymbol,
		), // getBaseBalance
		getTargetBalance(
			client, // client
			targetSymbol,
		), // getTargetBalance
		getFreeBalance(
			client, // client
			symbol,
		), // getFreeBalance
		getLockedBalance(
			client, // client
			symbol,
		), // getLockedBalance
		getCurrentPrice(
			client, // client
			symbol,
		), // getCurrentPrice
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
func depthsCreator(
	client *binance.Client,
	degree int,
	depthAPILimit depth_types.DepthAPILimit,
	depthsCallBack func(p *processor_types.Processor) func(d *depth_types.Depths) binance.WsDepthHandler,
	depthsErrHandler func(p *processor_types.Processor) func(d *depth_types.Depths) binance.ErrHandler) func(p *processor_types.Processor) processor_types.DepthConstructor {
	return func(p *processor_types.Processor) processor_types.DepthConstructor {
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
				p.GetSymbol(),
				spot_depth.DepthStreamCreator(
					callBack,
					errHandler),
				spot_depth.InitCreator(depthAPILimit, client))
		}
	}
} // depthsCreator

func ordersCreator(
	client *binance.Client,
	ordersCallBack func(p *processor_types.Processor) func(o *orders_types.Orders) binance.WsUserDataHandler,
	ordersErrHandler func(p *processor_types.Processor) func(o *orders_types.Orders) binance.ErrHandler,
	quit chan struct{}) func(p *processor_types.Processor) processor_types.OrdersConstructor {
	return func(p *processor_types.Processor) processor_types.OrdersConstructor {
		return func() *orders_types.Orders {
			var (
				callBack   func(*orders_types.Orders) binance.WsUserDataHandler
				errHandler func(*orders_types.Orders) binance.ErrHandler
			)
			if ordersCallBack != nil {
				callBack = spot_orders.CallBackCreator(ordersCallBack(p))
			} else {
				callBack = spot_orders.CallBackCreator()
			}
			if ordersErrHandler != nil {
				errHandler = spot_orders.WsErrorHandlerCreator(ordersErrHandler(p))
			} else {
				errHandler = spot_orders.WsErrorHandlerCreator()
			}
			return orders_types.New(
				p.GetSymbol(), // symbol
				spot_orders.UserDataStreamCreator(
					client,
					callBack,
					errHandler), // userDataStream
				spot_orders.CreateOrderCreator(
					client,
					int(float64(p.GetStepSizeExp())),
					int(float64(p.GetTickSizeExp()))), // createOrder
				spot_orders.GetOpenOrdersCreator(client),   // getOpenOrders
				spot_orders.GetAllOrdersCreator(client),    // getAllOrders
				spot_orders.GetOrderCreator(client),        // getOrder
				spot_orders.CancelOrderCreator(client),     // cancelOrder
				spot_orders.CancelAllOrdersCreator(client), // cancelAllOrders
				quit)
		}
	}
} // ordersCreator

func getBaseBalance(client *binance.Client, symbol string) processor_types.GetBaseBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.Free) + utils.ConvStrToFloat64(asset.Locked))
			}
		}
		return 0.0
	}
} // getBaseBalance
func getTargetBalance(client *binance.Client, symbol string) processor_types.GetTargetBalanceFunction {
	return func() items_types.QuantityType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.QuantityType(utils.ConvStrToFloat64(asset.Free) + utils.ConvStrToFloat64(asset.Locked))
			}
		}
		return 0.0
	}
} // getTargetBalance
func getFreeBalance(client *binance.Client, symbol string) processor_types.GetFreeBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.Free))
			}
		}
		return 0.0
	}
} // getFreeBalance
func getLockedBalance(client *binance.Client, symbol string) processor_types.GetLockedBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.Locked))
			}
		}
		return 0.0
	}
} // getLockedBalance
func getCurrentPrice(client *binance.Client, symbol string) processor_types.GetCurrentPriceFunction {
	return func() items_types.PriceType {
		price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get price: %v", err)
			return 0
		}
		return items_types.PriceType(utils.ConvStrToFloat64(price[0].Price))
	}
} // getCurrentPrice
