package processor

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"

	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/depths"
	futures_exchangeinfo "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_orders "github.com/fr0ster/go-trading-utils/binance/futures/orders"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchangeinfo_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	processor_types "github.com/fr0ster/go-trading-utils/types/processor"
)

func New(
	client *futures.Client,
	degree int,
	symbol string,
	depthAPILimit depth_types.DepthAPILimit,
	depthStreamLevel depth_types.DepthStreamLevel,
	depthStreamRate depth_types.DepthStreamRate,
	quits ...chan struct{},
) (pairProcessor *processor_types.Processor, err error) {
	var quit chan struct{}
	if len(quits) > 0 {
		quit = quits[0]
	} else {
		quit = make(chan struct{})
	}
	exchange := exchangeinfo_types.New(futures_exchangeinfo.InitCreator(client, degree, symbol))
	depths := depth_types.New(
		degree,
		symbol,
		futures_depth.DepthStreamCreator(
			depthStreamLevel,
			depthStreamRate,
			futures_depth.CallBackCreator(),
			futures_depth.WsErrorHandlerCreator()),
		futures_depth.InitCreator(depthAPILimit, client))
	symbolInfo := exchange.GetSymbols().GetSymbol(symbol)
	orders := orders_types.New(
		symbol, // symbol
		futures_orders.UserDataStreamCreator(
			client,
			futures_orders.CallBackCreator(),
			futures_orders.WsErrorHandlerCreator()), // userDataStream
		futures_orders.CreateOrderCreator(
			client,
			int(float64(symbolInfo.GetStepSize())),
			int(float64(symbolInfo.GetTickSizeExp()))), // createOrder
		futures_orders.GetOpenOrdersCreator(client),   // getOpenOrders
		futures_orders.GetAllOrdersCreator(client),    // getAllOrders
		futures_orders.GetOrderCreator(client),        // getOrder
		futures_orders.CancelOrderCreator(client),     // cancelOrder
		futures_orders.CancelAllOrdersCreator(client), // cancelAllOrders
		quit)
	account, _ := client.NewGetAccountService().Do(context.Background())
	pairProcessor, err = processor_types.New(
		quit,     // quit
		symbol,   // pair
		exchange, // exchange
		depths,   // depths
		orders,   // orders
		func() items_types.ValueType {
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
				}
			}
			return 0.0
		}, // getBaseBalance
		func() items_types.ValueType {
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetTargetSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
				}
			}
			return 0.0
		}, // getTargetBalance
		func() items_types.ValueType {
			for _, asset := range account.Assets {
				if asset.Asset == string(exchange.GetSymbol(symbol).GetBaseSymbol()) {
					return items_types.ValueType(utils.ConvStrToFloat64(asset.AvailableBalance))
				}
			}
			return 0.0
		}, // getFreeBalance
		func() items_types.ValueType {
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
		func() *futures.PositionRisk {
			risks, err := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
			if err == nil {
				return risks[0]
			}
			return nil
		}, // getPositionRisk
		func(p *processor_types.Processor) func(int) (*futures.SymbolLeverage, error) {
			return func(leverage int) (*futures.SymbolLeverage, error) {
				res, err := client.NewChangeLeverageService().Symbol(symbol).Leverage(leverage).Do(context.Background())
				return res, err
			}
		}, // setLeverage
		func(p *processor_types.Processor) func(pairs_types.MarginType) error {
			return func(marginType pairs_types.MarginType) error {
				return client.NewChangeMarginTypeService().Symbol(symbol).MarginType(futures.MarginType(marginType)).Do(context.Background())
			}
		}, // setMarginType
		func(p *processor_types.Processor) func(items_types.ValueType, int) error {
			return func(amountMargin items_types.ValueType, typeMargin int) error {
				return client.
					NewUpdatePositionMarginService().
					Symbol(symbol).
					Amount(utils.ConvFloat64ToStrDefault(float64(amountMargin))).
					Type(typeMargin).
					Do(context.Background())
			}
		}, // setPositionMargin
		func(p *processor_types.Processor) func(*futures.PositionRisk) error {
			return func(risk *futures.PositionRisk) (err error) {
				if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
					_, err = p.GetOrders().CreateOrder(
						orders_types.OrderType(futures.OrderTypeTakeProfitMarket),
						orders_types.SideType(futures.SideTypeBuy),
						orders_types.TimeInForceType(futures.TimeInForceTypeGTC),
						0, true, false, 0, 0, 0, 0)
				} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
					_, err = p.GetOrders().CreateOrder(
						orders_types.OrderType(futures.OrderTypeTakeProfitMarket),
						orders_types.SideType(futures.SideTypeBuy),
						orders_types.TimeInForceType(futures.TimeInForceTypeGTC),
						0, true, false, 0, 0, 0, 0)
				}
				return
			}
		}) // closePosition
	if err != nil {
		logrus.Errorf("Can't init pair: %v", err)
		close(quit)
		return
	}
	return
}
