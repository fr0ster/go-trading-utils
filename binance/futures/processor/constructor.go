package processor

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"

	futures_exchangeinfo "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchangeinfo_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
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
	leverage int,
	marginType types.MarginType,
	callbackRate items_types.PricePercentType,
	// depthAPILimit depth_types.DepthAPILimit,
	// depthStreamLevel depth_types.DepthStreamLevel,
	// depthStreamRate depth_types.DepthStreamRate,
	// ordersCallBack func(p *processor_types.Processor) func(o *orders_types.Orders) futures.WsUserDataHandler,
	// ordersErrHandler func(p *processor_types.Processor) func(o *orders_types.Orders) futures.ErrHandler,
	// depthsCallBack func(p *processor_types.Processor) func(d *depth_types.Depths) futures.WsDepthHandler,
	// depthsErrHandler func(p *processor_types.Processor) func(d *depth_types.Depths) futures.ErrHandler,
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
	symbolInfo := exchange.GetSymbol(symbol)
	baseSymbol := string(exchange.GetSymbol(symbol).GetBaseSymbol())
	targetSymbol := string(exchange.GetSymbol(symbol).GetTargetSymbol())
	pairProcessor, err = processor_types.New(
		quit,   // quit
		symbol, // pair
		// exchange, // exchange
		symbolInfo, // symbolInfo
		// depthsCreator(
		// 	client,           // client
		// 	degree,           // degree
		// 	depthAPILimit,    // depthAPILimit
		// 	depthStreamLevel, // depthStreamLevel
		// 	depthStreamRate,  // depthStreamRate
		// 	depthsCallBack,   // depthsCallBack
		// 	depthsErrHandler, // depthsErrHandler
		// ), // depthsCreator
		// ordersCreator(
		// 	client,           // client
		// 	ordersCallBack,   // ordersCallBack
		// 	ordersErrHandler, // ordersCreator
		// 	quit,             // ordersCreator
		// ), // ordersCreator
		getBaseBalance(
			client,     // client
			baseSymbol, // symbol
		), // getBaseBalance
		getTargetBalance(
			client,       // client
			targetSymbol, // symbol
		), // getTargetBalance
		getFreeBalance(
			client,     // client
			baseSymbol, // symbol
		), // getFreeBalance
		getLockedBalance(
			client,     // client
			baseSymbol, // symbol
		), // getLockedBalance
		getCurrentPrice(client, symbol), // getCurrentPrice
		getPositionRisk(client),         // getPositionRisk
		func() int {
			return leverage
		}, // getLeverage
		setLeverage(client), // setLeverage
		func() types.MarginType {
			return marginType
		}, // getMarginType
		setMarginType(client),     // setMarginType
		setPositionMargin(client), // setPositionMargin
		// closePosition(),           // closePosition
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
	return
}

// func depthsCreator(
// 	client *futures.Client,
// 	degree int,
// 	depthAPILimit depth_types.DepthAPILimit,
// 	depthStreamLevel depth_types.DepthStreamLevel,
// 	depthStreamRate depth_types.DepthStreamRate,
// 	depthsCallBack func(p *processor_types.Processor) func(d *depth_types.Depths) futures.WsDepthHandler,
// 	depthsErrHandler func(p *processor_types.Processor) func(d *depth_types.Depths) futures.ErrHandler) func(p *processor_types.Processor) processor_types.DepthConstructor {
// 	return func(p *processor_types.Processor) processor_types.DepthConstructor {
// 		return func() *depth_types.Depths {
// 			var (
// 				callBack   func(d *depth_types.Depths) futures.WsDepthHandler
// 				errHandler func(*depth_types.Depths) futures.ErrHandler
// 			)
// 			if depthsCallBack != nil {
// 				callBack = futures_depth.CallBackCreator(depthsCallBack(p))
// 			} else {
// 				callBack = futures_depth.CallBackCreator()
// 			}
// 			if depthsErrHandler != nil {
// 				errHandler = futures_depth.WsErrorHandlerCreator(depthsErrHandler(p))
// 			} else {
// 				errHandler = futures_depth.WsErrorHandlerCreator()
// 			}
// 			return depth_types.New(
// 				degree,
// 				p.GetSymbol(),
// 				futures_depth.DepthStreamCreator(
// 					depthStreamLevel,
// 					depthStreamRate,
// 					callBack,
// 					errHandler),
// 				futures_depth.InitCreator(depthAPILimit, client))
// 		}
// 	}
// } // depthsCreator

// func ordersCreator(
// 	client *futures.Client,
// 	ordersCallBack func(p *processor_types.Processor) func(o *orders_types.Orders) futures.WsUserDataHandler,
// 	ordersErrHandler func(p *processor_types.Processor) func(o *orders_types.Orders) futures.ErrHandler,
// 	quit chan struct{}) func(p *processor_types.Processor) processor_types.OrdersConstructor {
// 	return func(p *processor_types.Processor) processor_types.OrdersConstructor {
// 		return func() *orders_types.Orders {
// 			var (
// 				callBack   func(*orders_types.Orders) futures.WsUserDataHandler
// 				errHandler func(*orders_types.Orders) futures.ErrHandler
// 			)
// 			if ordersCallBack != nil {
// 				callBack = futures_orders.CallBackCreator(ordersCallBack(p))
// 			} else {
// 				callBack = futures_orders.CallBackCreator()
// 			}
// 			if ordersErrHandler != nil {
// 				errHandler = futures_orders.WsErrorHandlerCreator(ordersErrHandler(p))
// 			} else {
// 				errHandler = futures_orders.WsErrorHandlerCreator()
// 			}
// 			return orders_types.New(
// 				p.GetSymbol(), // symbol
// 				futures_orders.UserDataStreamCreator(
// 					client,
// 					callBack,
// 					errHandler), // userDataStream
// 				futures_orders.CreateOrderCreator(
// 					client,
// 					int(float64(p.GetStepSizeExp())),
// 					int(float64(p.GetTickSizeExp()))), // createOrder
// 				futures_orders.GetOpenOrdersCreator(client),   // getOpenOrders
// 				futures_orders.GetAllOrdersCreator(client),    // getAllOrders
// 				futures_orders.GetOrderCreator(client),        // getOrder
// 				futures_orders.CancelOrderCreator(client),     // cancelOrder
// 				futures_orders.CancelAllOrdersCreator(client), // cancelAllOrders
// 				quit)
// 		}
// 	}
// } // ordersCreator

func getBaseBalance(client *futures.Client, symbol string) processor_types.GetBaseBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
			}
		}
		return 0.0
	}
} // getBaseBalance
func getTargetBalance(client *futures.Client, symbol string) processor_types.GetTargetBalanceFunction {
	return func() items_types.QuantityType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.QuantityType(utils.ConvStrToFloat64(asset.WalletBalance))
			}
		}
		return 0.0
	}
} // getTargetBalance
func getFreeBalance(client *futures.Client, symbol string) processor_types.GetFreeBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.AvailableBalance))
			}
		}
		return 0.0
	}
} // getFreeBalance
func getLockedBalance(client *futures.Client, symbol string) processor_types.GetLockedBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance) - utils.ConvStrToFloat64(asset.AvailableBalance))
			}
		}
		return 0.0
	}
} // getLockedBalance
func getCurrentPrice(client *futures.Client, symbol string) processor_types.GetCurrentPriceFunction {
	return func() items_types.PriceType {
		price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get price: %v", err)
			return 0
		}
		return items_types.PriceType(utils.ConvStrToFloat64(price[0].Price))
	}
} // getCurrentPrice
func getPositionRisk(client *futures.Client) func(*processor_types.Processor) processor_types.GetPositionRiskFunction {
	return func(p *processor_types.Processor) processor_types.GetPositionRiskFunction {
		return func() *futures.PositionRisk {
			risks, err := client.NewGetPositionRiskService().Symbol(p.GetSymbol()).Do(context.Background())
			if err == nil {
				return risks[0]
			}
			return &futures.PositionRisk{}
		}
	}
} // getPositionRisk
func setLeverage(client *futures.Client) func(p *processor_types.Processor) processor_types.SetLeverageFunction {
	return func(p *processor_types.Processor) processor_types.SetLeverageFunction {
		return func(leverage int) (Leverage int, MaxNotionalValue string, Symbol string, err error) {
			var res *futures.SymbolLeverage
			res, err = client.NewChangeLeverageService().Symbol(p.GetSymbol()).Leverage(leverage).Do(context.Background())
			Leverage = res.Leverage
			MaxNotionalValue = res.MaxNotionalValue
			Symbol = res.Symbol
			return
		}
	}
} // setLeverage
func setMarginType(client *futures.Client) func(p *processor_types.Processor) processor_types.SetMarginTypeFunction {
	return func(p *processor_types.Processor) processor_types.SetMarginTypeFunction {
		return func(marginType types.MarginType) error {
			return client.NewChangeMarginTypeService().Symbol(p.GetSymbol()).MarginType(futures.MarginType(marginType)).Do(context.Background())
		}
	}
} // setMarginType
func setPositionMargin(client *futures.Client) func(p *processor_types.Processor) processor_types.SetPositionMarginFunction {
	return func(p *processor_types.Processor) processor_types.SetPositionMarginFunction {
		return func(amountMargin items_types.ValueType, typeMargin int) error {
			return client.
				NewUpdatePositionMarginService().
				Symbol(p.GetSymbol()).
				Amount(utils.ConvFloat64ToStrDefault(float64(amountMargin))).
				Type(typeMargin).
				Do(context.Background())
		}
	}
} // setPositionMargin
// func closePosition(debug ...*futures.PositionRisk) func(p *processor_types.Processor) processor_types.ClosePositionFunction {
// 	return func(p *processor_types.Processor) processor_types.ClosePositionFunction {
// 		return func() (err error) {
// 			risk := p.GetPositionRisk(debug...)
// 			if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
// 				_, err = p.GetOrders().CreateOrder(
// 					types.OrderType(futures.OrderTypeTakeProfitMarket),
// 					types.SideType(futures.SideTypeBuy),
// 					types.TimeInForceType(futures.TimeInForceTypeGTC),
// 					0, true, false, 0, 0, 0, 0)
// 			} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
// 				_, err = p.GetOrders().CreateOrder(
// 					types.OrderType(futures.OrderTypeTakeProfitMarket),
// 					types.SideType(futures.SideTypeSell),
// 					types.TimeInForceType(futures.TimeInForceTypeGTC),
// 					0, true, false, 0, 0, 0, 0)
// 			}
// 			return
// 		}
// 	}
// } // closePosition
