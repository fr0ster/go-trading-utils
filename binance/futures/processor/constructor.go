package processor

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"

	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/depths"
	futures_exchangeinfo "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_orders "github.com/fr0ster/go-trading-utils/binance/futures/orders"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchangeinfo_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	processor_types "github.com/fr0ster/go-trading-utils/types/processor"
)

func New(
	client *futures.Client,
	degree int,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpAndLowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	callbackRate items_types.PricePercentType,
	depthAPILimit depth_types.DepthAPILimit,
	depthStreamLevel depth_types.DepthStreamLevel,
	depthStreamRate depth_types.DepthStreamRate,
	ordersCallBack func(p *processor_types.Processor) func(o *orders_types.Orders) futures.WsUserDataHandler,
	depthsCallBack func(p *processor_types.Processor) func(d *depth_types.Depths) futures.WsDepthHandler,
	debug bool,
	quits ...chan struct{},
) (pairProcessor *processor_types.Processor, err error) {
	var quit chan struct{}
	if len(quits) > 0 {
		quit = quits[0]
	} else {
		quit = make(chan struct{})
	}
	exchange := exchangeinfo_types.New(futures_exchangeinfo.InitCreator(client, degree, symbol))
	symbolInfo := exchange.GetSymbols().GetSymbol(symbol)
	pairProcessor, err = processor_types.New(
		quit,     // quit
		symbol,   // pair
		exchange, // exchange
		func(p *processor_types.Processor) *depth_types.Depths {
			return depth_types.New(
				degree,
				symbol,
				futures_depth.DepthStreamCreator(
					depthStreamLevel,
					depthStreamRate,
					futures_depth.CallBackCreator(depthsCallBack(p)),
					futures_depth.WsErrorHandlerCreator()),
				futures_depth.InitCreator(depthAPILimit, client))
		}, // depthsCreator
		func(p *processor_types.Processor) *orders_types.Orders {
			return orders_types.New(
				symbol, // symbol
				futures_orders.UserDataStreamCreator(
					client,
					futures_orders.CallBackCreator(ordersCallBack(p)),
					futures_orders.WsErrorHandlerCreator()), // userDataStream
				futures_orders.CreateOrderCreator(
					client,
					int(float64(symbolInfo.GetStepSize())),
					int(float64(symbolInfo.GetTickSize()))), // createOrder
				futures_orders.GetOpenOrdersCreator(client),   // getOpenOrders
				futures_orders.GetAllOrdersCreator(client),    // getAllOrders
				futures_orders.GetOrderCreator(client),        // getOrder
				futures_orders.CancelOrderCreator(client),     // cancelOrder
				futures_orders.CancelAllOrdersCreator(client), // cancelAllOrders
				quit)
		}, // orders
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
				}
			}
			return 0.0
		}, // getBaseBalance
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetTargetSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
				}
			}
			return 0.0
		}, // getTargetBalance
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.AvailableBalance))
				}
			}
			return 0.0
		}, // getFreeBalance
		func() items_types.ValueType {
			account, _ := client.NewGetAccountService().Do(context.Background())
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance) - utils.ConvStrToFloat64(asset.AvailableBalance))
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
		func(*processor_types.Processor) processor_types.GetPositionRiskFunction {
			return func() *futures.PositionRisk {
				risks, err := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
				if err == nil {
					return risks[0]
				}
				return nil
			}
		}, // getPositionRisk
		func(p *processor_types.Processor) processor_types.SetLeverageFunction {
			return func(leverage int) (Leverage int, MaxNotionalValue string, Symbol string, err error) {
				var res *futures.SymbolLeverage
				res, err = client.NewChangeLeverageService().Symbol(symbol).Leverage(leverage).Do(context.Background())
				Leverage = res.Leverage
				MaxNotionalValue = res.MaxNotionalValue
				Symbol = res.Symbol
				return
			}
		}, // setLeverage
		func(p *processor_types.Processor) processor_types.SetMarginTypeFunction {
			return func(marginType types.MarginType) error {
				return client.NewChangeMarginTypeService().Symbol(symbol).MarginType(futures.MarginType(marginType)).Do(context.Background())
			}
		}, // setMarginType
		func(p *processor_types.Processor) processor_types.SetPositionMarginFunction {
			return func(amountMargin items_types.ValueType, typeMargin int) error {
				return client.
					NewUpdatePositionMarginService().
					Symbol(symbol).
					Amount(utils.ConvFloat64ToStrDefault(float64(amountMargin))).
					Type(typeMargin).
					Do(context.Background())
			}
		}, // setPositionMargin
		func(p *processor_types.Processor) processor_types.ClosePositionFunction {
			return func() (err error) {
				risk := p.GetPositionRisk()
				if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
					_, err = p.GetOrders().CreateOrder(
						types.OrderType(futures.OrderTypeTakeProfitMarket),
						types.SideType(futures.SideTypeBuy),
						types.TimeInForceType(futures.TimeInForceTypeGTC),
						0, true, false, 0, 0, 0, 0)
				} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
					_, err = p.GetOrders().CreateOrder(
						types.OrderType(futures.OrderTypeTakeProfitMarket),
						types.SideType(futures.SideTypeBuy),
						types.TimeInForceType(futures.TimeInForceTypeGTC),
						0, true, false, 0, 0, 0, 0)
				}
				return
			}
		}, // closePosition
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
