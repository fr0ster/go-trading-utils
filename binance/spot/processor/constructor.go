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
	depthsCallBack func(p *processor_types.Processor) func(d *depth_types.Depths) binance.WsDepthHandler,
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
		func(p *processor_types.Processor) *depth_types.Depths {
			return depth_types.New(
				degree,
				symbol,
				spot_depth.DepthStreamCreator(
					spot_depth.CallBackCreator(depthsCallBack(p)),
					spot_depth.WsErrorHandlerCreator()),
				spot_depth.InitCreator(depthAPILimit, client))
		}, // depthsCreator
		func(p *processor_types.Processor) *orders_types.Orders {
			return orders_types.New(
				symbol, // symbol
				spot_orders.UserDataStreamCreator(
					client,
					spot_orders.CallBackCreator(ordersCallBack(p)),
					spot_orders.WsErrorHandlerCreator()), // userDataStream
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
		nil, // setLeverage
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
