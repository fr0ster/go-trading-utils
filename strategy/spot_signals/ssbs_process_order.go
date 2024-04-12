package spot_signals

import (
	"context"
	"math"
	_ "net/http/pprof"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	utils "github.com/fr0ster/go-trading-utils/utils"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

func ProcessBuyOrder(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	orderType binance.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *depth_types.DepthItemType,
	stopBuy chan bool,
	stopEvent chan os.Signal) (startBuyOrderEvent chan *binance.CreateOrderResponse) {
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).PriceFilter().TickSize)))
	)
	go func() {
		var order *binance.CreateOrderResponse
		var err error
		for {
			select {
			case <-stopBuy:
				return
			case <-stopEvent:
				return
			case params := <-buyEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				service :=
					client.NewCreateOrderService().
						Symbol(string(binance.SymbolType((*pair).GetPair()))).
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
						(*pair).GetPair(), binance.SideTypeBuy, params.Quantity, params.Price)
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
						(*pair).SetBuyQuantity((*pair).GetBuyQuantity() + fillQuantity)
						(*pair).SetBuyValue((*pair).GetBuyValue() + fillQuantity*fillPrice)
						(*pair).CalcMiddlePrice()
						(*pair).AddCommission(fill)
					}
					config.Save()
				}
			}
		}
	}()
	return
}

func ProcessSellOrder(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	orderType binance.OrderType,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *depth_types.DepthItemType,
	stopSell chan bool,
	stopEvent chan os.Signal) (startSellOrderEvent chan *binance.CreateOrderResponse) {
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).PriceFilter().TickSize)))
	)
	startSellOrderEvent = make(chan *binance.CreateOrderResponse)
	go func() {
		var order *binance.CreateOrderResponse
		var err error
		for {
			select {
			case <-stopSell:
				return
			case <-stopEvent:
				return
			case params := <-sellEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				service :=
					client.NewCreateOrderService().
						Symbol(string(binance.SymbolType((*pair).GetPair()))).
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
						(*pair).GetPair(), binance.SideTypeSell, params.Quantity, params.Price)
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
						(*pair).SetBuyQuantity((*pair).GetBuyQuantity() + fillQuantity)
						(*pair).SetBuyValue((*pair).GetBuyValue() + fillQuantity*fillPrice)
						(*pair).CalcMiddlePrice()
						(*pair).AddCommission(fill)
					}
					config.Save()
				}
			}
		}
	}()
	return
}

func ProcessAfterBuyOrder(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *depth_types.DepthItemType,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	stopBuy chan bool,
	stopSell chan bool,
	startBuyOrderEvent chan *binance.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-stopBuy:
				return
			case <-stopSell:
				return
			case <-stopEvent:
				return
			case order := <-startBuyOrderEvent:
				if order != nil {
					for {
						orderEvent := <-orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
								(*pair).SetBuyQuantity((*pair).GetBuyQuantity() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
								(*pair).SetBuyValue((*pair).GetBuyValue() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
								(*pair).CalcMiddlePrice()
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
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *depth_types.DepthItemType,
	stopBuy chan bool,
	stopSell chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	startSellOrderEvent chan *binance.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-stopBuy:
				return
			case <-stopSell:
				return
			case <-stopEvent:
				return
			case order := <-startSellOrderEvent:
				if order != nil {
					for {
						orderEvent := <-orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
								(*pair).SetSellQuantity((*pair).GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
								(*pair).SetSellValue((*pair).GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
								(*pair).CalcMiddlePrice()
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

func ProcessAfterOrder(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	sellEvent chan *depth_types.DepthItemType,
	stopProcess chan bool,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	order *binance.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-stopProcess:
				return
			case <-stopEvent:
				return
			case orderEvent := <-orderStatusEvent:
				logrus.Debug("Order status changed")
				if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
					if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
						orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
						(*pair).SetSellQuantity((*pair).GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
						(*pair).SetSellValue((*pair).GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
						(*pair).CalcMiddlePrice()
						(*pair).SetStage(pairs_types.PositionClosedStage)
						config.Save()
						break
					}
				}
			}
		}
	}()
}
