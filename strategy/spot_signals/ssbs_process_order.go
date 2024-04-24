package spot_signals

import (
	"context"
	"log"
	"math"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	utils "github.com/fr0ster/go-trading-utils/utils"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func ProcessBuyOrder(
	config *config_types.ConfigFile,
	client *binance.Client,
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	orderType binance.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *pair_price_types.PairPrice,
	stopBuy chan bool,
	stopEvent chan os.Signal) (startBuyOrderEvent chan *binance.CreateOrderResponse) {
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
		var order *binance.CreateOrderResponse
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
						Symbol(string(binance.SymbolType(pair.GetPair()))).
						Type(orderType).
						Side(binance.SideTypeBuy).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
				if orderType == binance.OrderTypeMarket {
					order, err = service.Do(context.Background())
				} else if orderType == binance.OrderTypeLimit {
					order, err = service.
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
				}
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pair.GetPair(), binance.SideTypeBuy, params.Quantity, params.Price)
					stopEvent <- os.Interrupt
					return
				}
				minuteOrderLimit.Limit++
				dayOrderLimit.Limit++
				if order.Status == binance.OrderStatusTypeNew {
					startBuyOrderEvent <- order
				} else {
					for _, fill := range order.Fills {
						fillPrice := utils.ConvStrToFloat64(fill.Price)
						fillQuantity := utils.ConvStrToFloat64(fill.Quantity)
						pair.SetBuyQuantity(pair.GetBuyQuantity() + fillQuantity)
						pair.SetBuyValue(pair.GetBuyValue() + fillQuantity*fillPrice)
						pair.CalcMiddlePrice()
						pair.AddCommission(fill)
					}
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
	client *binance.Client,
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	orderType binance.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *pair_price_types.PairPrice,
	stopSell chan bool,
	stopEvent chan os.Signal) (startSellOrderEvent chan *binance.CreateOrderResponse) {
	symbol, err := (*pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	startSellOrderEvent = make(chan *binance.CreateOrderResponse)
	go func() {
		var order *binance.CreateOrderResponse
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
						Symbol(string(binance.SymbolType(pair.GetPair()))).
						Type(binance.OrderTypeLimit).
						Side(binance.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
				if orderType == binance.OrderTypeMarket {
					order, err = service.Do(context.Background())
				} else if orderType == binance.OrderTypeLimit {
					order, err = service.
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
				}
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pair.GetPair(), binance.SideTypeSell, params.Quantity, params.Price)
					stopEvent <- os.Interrupt
					return
				}
				minuteOrderLimit.Limit++
				dayOrderLimit.Limit++
				if order.Status == binance.OrderStatusTypeNew {
					startSellOrderEvent <- order
				} else {
					for _, fill := range order.Fills {
						fillPrice := utils.ConvStrToFloat64(fill.Price)
						fillQuantity := utils.ConvStrToFloat64(fill.Quantity)
						pair.SetBuyQuantity(pair.GetBuyQuantity() + fillQuantity)
						pair.SetBuyValue(pair.GetBuyValue() + fillQuantity*fillPrice)
						pair.CalcMiddlePrice()
						pair.AddCommission(fill)
					}
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
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	orderType binance.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *pair_price_types.PairPrice,
	stopProcess chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent) (startBuyOrderEvent chan *binance.CreateOrderResponse) {
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
						Symbol(string(binance.SymbolType(pair.GetPair()))).
						Type(binance.OrderTypeTakeProfit).
						Side(binance.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pair.GetPair(), binance.SideTypeBuy, params.Quantity, params.Price)
					stopEvent <- os.Interrupt
					return
				}
				minuteOrderLimit.Limit++
				dayOrderLimit.Limit++
				if order.Status == binance.OrderStatusTypeNew {
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
					for _, fill := range order.Fills {
						fillPrice := utils.ConvStrToFloat64(fill.Price)
						fillQuantity := utils.ConvStrToFloat64(fill.Quantity)
						pair.SetBuyQuantity(pair.GetBuyQuantity() + fillQuantity)
						pair.SetBuyValue(pair.GetBuyValue() + fillQuantity*fillPrice)
						pair.CalcMiddlePrice()
						pair.AddCommission(fill)
					}
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
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *pair_price_types.PairPrice,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	stopBuy chan bool,
	stopSell chan bool,
	startBuyOrderEvent chan *binance.CreateOrderResponse) {
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
						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
								pair.SetBuyQuantity(pair.GetBuyQuantity() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
								pair.SetBuyValue(pair.GetBuyValue() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
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
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *pair_price_types.PairPrice,
	stopBuy chan bool,
	stopSell chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	startSellOrderEvent chan *binance.CreateOrderResponse) {
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
						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
								pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
								pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
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
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	stopProcess chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	order *binance.CreateOrderResponse) (orderExecuted chan bool) {
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
				if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
					if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
						orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
						pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
						pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
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
