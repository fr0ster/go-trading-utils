package futures_signals

import (
	"context"
	"log"
	"math"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	utils "github.com/fr0ster/go-trading-utils/utils"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

func ProcessBuyOrder(
	config *config_types.ConfigFile,
	client *futures.Client,
	account account_interfaces.Accounts,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	orderType futures.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *depth_types.DepthItemType,
	stopBuy chan bool,
	stopEvent chan os.Signal) (startBuyOrderEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	go func() {
		var order *futures.CreateOrderResponse
		for {
			select {
			case <-stopBuy:
				stopBuy <- true
				return
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case params := <-buyEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				if params.Price == 0 || params.Quantity == 0 {
					continue
				}
				targetBalance, err := GetTargetBalance(account, pair)
				if err != nil {
					logrus.Errorf("Can't get %s asset: %v", pair.GetBaseSymbol(), err)
					stopEvent <- os.Interrupt
					return
				}
				if targetBalance > pair.GetLimitInputIntoPosition() || targetBalance > pair.GetLimitOutputOfPosition() {
					logrus.Warnf("We'd buy %s lots of %s, but we have not enough %s",
						pair.GetPair(), pair.GetBaseSymbol(), pair.GetBaseSymbol())
					continue
				}
				service :=
					client.NewCreateOrderService().
						Symbol(string(futures.SymbolType(pair.GetPair()))).
						Type(orderType).
						Side(futures.SideTypeBuy).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
				if orderType == futures.OrderTypeMarket {
					order, err = service.Do(context.Background())
				} else if orderType == futures.OrderTypeLimit {
					order, err = service.
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
				}
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pair.GetPair(), futures.SideTypeBuy, params.Quantity, params.Price)
					stopEvent <- os.Interrupt
					return
				}
				minuteOrderLimit.Limit++
				dayOrderLimit.Limit++
				if order.Status == futures.OrderStatusTypeNew {
					startBuyOrderEvent <- order
				} else {
					fillPrice := utils.ConvStrToFloat64(order.Price)
					fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
					pair.SetBuyQuantity(pair.GetBuyQuantity() + fillQuantity)
					pair.SetBuyValue(pair.GetBuyValue() + fillQuantity*fillPrice)
					pair.CalcMiddlePrice()
					config.Save()
				}
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func ProcessSellOrder(
	config *config_types.ConfigFile,
	client *futures.Client,
	account account_interfaces.Accounts,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	orderType futures.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *depth_types.DepthItemType,
	stopSell chan bool,
	stopEvent chan os.Signal) (startSellOrderEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	startSellOrderEvent = make(chan *futures.CreateOrderResponse)
	go func() {
		var order *futures.CreateOrderResponse
		// var err error
		for {
			select {
			case <-stopSell:
				stopSell <- true
				return
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case params := <-sellEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				if params.Price == 0 || params.Quantity == 0 {
					continue
				}
				targetBalance, err := GetTargetBalance(account, pair)
				if err != nil {
					logrus.Errorf("Can't get %s asset: %v", pair.GetBaseSymbol(), err)
					stopEvent <- os.Interrupt
					return
				}
				if targetBalance < params.Price*params.Quantity {
					logrus.Warnf("We don't have enough %s to sell %s lots of %s",
						pair.GetPair(), pair.GetBaseSymbol(), pair.GetBaseSymbol())
					continue
				}
				service :=
					client.NewCreateOrderService().
						Symbol(string(futures.SymbolType(pair.GetPair()))).
						Type(futures.OrderTypeLimit).
						Side(futures.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
				if orderType == futures.OrderTypeMarket {
					order, err = service.Do(context.Background())
				} else if orderType == futures.OrderTypeLimit {
					order, err = service.
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
				}
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pair.GetPair(), futures.SideTypeSell, params.Quantity, params.Price)
					stopEvent <- os.Interrupt
					return
				}
				minuteOrderLimit.Limit++
				dayOrderLimit.Limit++
				if order.Status == futures.OrderStatusTypeNew {
					startSellOrderEvent <- order
				} else {
					fillPrice := utils.ConvStrToFloat64(order.Price)
					fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
					pair.SetBuyQuantity(pair.GetBuyQuantity() + fillQuantity)
					pair.SetBuyValue(pair.GetBuyValue() + fillQuantity*fillPrice)
					pair.CalcMiddlePrice()
					config.Save()
				}
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func ProcessSellTakeProfitOrder(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	orderType futures.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *depth_types.DepthItemType,
	stopProcess chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *futures.WsUserDataEvent) (startBuyOrderEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	go func() {
		for {
			select {
			case <-stopProcess:
				stopProcess <- true
				return
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case params := <-sellEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				order, err :=
					client.NewCreateOrderService().
						Symbol(string(futures.SymbolType(pair.GetPair()))).
						Type(futures.OrderTypeTakeProfit).
						Side(futures.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pair.GetPair(), futures.SideTypeBuy, params.Quantity, params.Price)
					stopEvent <- os.Interrupt
					return
				}
				minuteOrderLimit.Limit++
				dayOrderLimit.Limit++
				if order.Status == futures.OrderStatusTypeNew {
					orderExecutionGuard := OrderExecutionGuard(
						config,
						client,
						pair,
						pairInfo,
						minuteOrderLimit,
						dayOrderLimit,
						minuteRawRequestLimit,
						stopProcess,
						stopEvent,
						orderStatusEvent,
						order)
					<-orderExecutionGuard
					startBuyOrderEvent <- order
				} else {
					fillPrice := utils.ConvStrToFloat64(order.Price)
					fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
					pair.SetBuyQuantity(pair.GetBuyQuantity() + fillQuantity)
					pair.SetBuyValue(pair.GetBuyValue() + fillQuantity*fillPrice)
					pair.CalcMiddlePrice()
					config.Save()
				}
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func ProcessAfterBuyOrder(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *depth_types.DepthItemType,
	stopEvent chan os.Signal,
	orderStatusEvent chan *futures.WsUserDataEvent,
	stopBuy chan bool,
	stopSell chan bool,
	startBuyOrderEvent chan *futures.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-stopBuy:
				stopBuy <- true
				return
			case <-stopSell:
				stopSell <- true
				return
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case order := <-startBuyOrderEvent:
				if order != nil {
					for {
						orderEvent := <-orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderTradeUpdate.ID == order.OrderID || orderEvent.OrderTradeUpdate.ClientOrderID == order.ClientOrderID {
							if orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
								orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled {
								pair.SetBuyQuantity(pair.GetBuyQuantity() - utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty))
								pair.SetBuyValue(pair.GetBuyValue() - utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty)*utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledPrice))
								pair.CalcMiddlePrice()
								config.Save()
								break
							}
						}
					}
				}
			}
		}
	}()
}

func ProcessAfterSellOrder(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *depth_types.DepthItemType,
	stopBuy chan bool,
	stopSell chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *futures.WsUserDataEvent,
	startSellOrderEvent chan *futures.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-stopBuy:
				stopBuy <- true
				return
			case <-stopSell:
				stopSell <- true
				return
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case order := <-startSellOrderEvent:
				if order != nil {
					for {
						orderEvent := <-orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderTradeUpdate.ID == order.OrderID || orderEvent.OrderTradeUpdate.ClientOrderID == order.ClientOrderID {
							if orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
								orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled {
								pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty))
								pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty)*utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledPrice))
								pair.CalcMiddlePrice()
								config.Save()
								break
							}
						}
					}
				}
			}
		}
	}()
}

func OrderExecutionGuard(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	stopProcess chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *futures.WsUserDataEvent,
	order *futures.CreateOrderResponse) (orderExecuted chan bool) {
	orderExecuted = make(chan bool)
	go func() {
		for {
			select {
			case <-stopProcess:
				stopProcess <- true
				return
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case orderEvent := <-orderStatusEvent:
				logrus.Debug("Order status changed")
				if orderEvent.OrderTradeUpdate.ID == order.OrderID || orderEvent.OrderTradeUpdate.ClientOrderID == order.ClientOrderID {
					if orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
						orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled {
						pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty))
						pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty)*utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledPrice))
						pair.CalcMiddlePrice()
						pair.SetStage(pairs_types.PositionClosedStage)
						config.Save()
						orderExecuted <- true
						return
					}
				}
			}
		}
	}()
	return
}
