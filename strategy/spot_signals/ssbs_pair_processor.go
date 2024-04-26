package spot_signals

import (
	"context"
	"log"
	"math"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	utils "github.com/fr0ster/go-trading-utils/utils"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairProcessor struct {
		config                *config_types.ConfigFile
		client                *binance.Client
		pair                  pairs_interfaces.Pairs
		exchangeInfo          *exchange_types.ExchangeInfo
		account               *spot_account.Account
		orderType             binance.OrderType
		buyEvent              chan *pair_price_types.PairPrice
		sellEvent             chan *pair_price_types.PairPrice
		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits
		stop                  chan os.Signal
		limitsOut             chan bool
		stopBuy               chan bool
		stopSell              chan bool
		stopAfterProcess      chan bool
		orderExecuted         chan bool
		orderStatusEvent      chan *binance.WsUserDataEvent
		userDataStream4Order  *spot_streams.UserDataStream
		pairInfo              *symbol_types.SpotSymbol
		degree                int
	}
)

func (pp *PairProcessor) ProcessBuyOrder() (startBuyOrderEvent chan *binance.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
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
			case <-pp.stopBuy:
				pp.stopBuy <- true
				return
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case params := <-pp.buyEvent:
				if pp.minuteOrderLimit.Limit == 0 || pp.dayOrderLimit.Limit == 0 || pp.minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				if params.Price == 0 || params.Quantity == 0 {
					continue
				}
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Errorf("Can't get %s asset: %v", pp.pair.GetBaseSymbol(), err)
					pp.stop <- os.Interrupt
					return
				}
				if targetBalance > pp.pair.GetLimitInputIntoPosition() || targetBalance > pp.pair.GetLimitOutputOfPosition() {
					logrus.Warnf("We'd buy %s lots of %s, but we have not enough %s",
						pp.pair.GetPair(), pp.pair.GetBaseSymbol(), pp.pair.GetBaseSymbol())
					continue
				}
				service :=
					pp.client.NewCreateOrderService().
						Symbol(string(binance.SymbolType(pp.pair.GetPair()))).
						Type(pp.orderType).
						Side(binance.SideTypeBuy).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
				if pp.orderType == binance.OrderTypeMarket {
					order, err = service.Do(context.Background())
				} else if pp.orderType == binance.OrderTypeLimit {
					order, err = service.
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
				}
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pp.pair.GetPair(), binance.SideTypeBuy, params.Quantity, params.Price)
					pp.stop <- os.Interrupt
					return
				}
				pp.minuteOrderLimit.Limit++
				pp.dayOrderLimit.Limit++
				if order.Status == binance.OrderStatusTypeNew {
					startBuyOrderEvent <- order
				} else {
					for _, fill := range order.Fills {
						fillPrice := utils.ConvStrToFloat64(fill.Price)
						fillQuantity := utils.ConvStrToFloat64(fill.Quantity)
						pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
						pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
						pp.pair.CalcMiddlePrice()
						pp.pair.AddCommission(fill)
					}
					pp.config.Save()
				}
			}
			time.Sleep(pp.pair.GetSleepingTime())
		}
	}()
	return
}

func (pp *PairProcessor) ProcessSellOrder() (startSellOrderEvent chan *binance.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
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
			case <-pp.stopSell:
				pp.stopSell <- true
				return
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case params := <-pp.sellEvent:
				if pp.minuteOrderLimit.Limit == 0 || pp.dayOrderLimit.Limit == 0 || pp.minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				if params.Price == 0 || params.Quantity == 0 {
					continue
				}
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Errorf("Can't get %s asset: %v", pp.pair.GetBaseSymbol(), err)
					pp.stop <- os.Interrupt
					return
				}
				if targetBalance < params.Price*params.Quantity {
					logrus.Warnf("We don't have enough %s to sell %s lots of %s",
						pp.pair.GetPair(), pp.pair.GetBaseSymbol(), pp.pair.GetBaseSymbol())
					continue
				}
				service :=
					pp.client.NewCreateOrderService().
						Symbol(string(binance.SymbolType(pp.pair.GetPair()))).
						Type(binance.OrderTypeLimit).
						Side(binance.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
				if pp.orderType == binance.OrderTypeMarket {
					order, err = service.Do(context.Background())
				} else if pp.orderType == binance.OrderTypeLimit {
					order, err = service.
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
				}
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pp.pair.GetPair(), binance.SideTypeSell, params.Quantity, params.Price)
					pp.stop <- os.Interrupt
					return
				}
				pp.minuteOrderLimit.Limit++
				pp.dayOrderLimit.Limit++
				if order.Status == binance.OrderStatusTypeNew {
					startSellOrderEvent <- order
				} else {
					for _, fill := range order.Fills {
						fillPrice := utils.ConvStrToFloat64(fill.Price)
						fillQuantity := utils.ConvStrToFloat64(fill.Quantity)
						pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
						pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
						pp.pair.CalcMiddlePrice()
						pp.pair.AddCommission(fill)
					}
					pp.config.Save()
				}
			}
			time.Sleep(pp.pair.GetSleepingTime())
		}
	}()
	return
}

func (pp *PairProcessor) ProcessSellTakeProfitOrder() (startBuyOrderEvent chan *binance.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
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
			case <-pp.stopAfterProcess:
				pp.stopAfterProcess <- true
				return
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case params := <-pp.sellEvent:
				if pp.minuteOrderLimit.Limit == 0 || pp.dayOrderLimit.Limit == 0 || pp.minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				order, err :=
					pp.client.NewCreateOrderService().
						Symbol(string(binance.SymbolType(pp.pair.GetPair()))).
						Type(binance.OrderTypeTakeProfit).
						Side(binance.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
				if err != nil {
					logrus.Errorf("Can't create order: %v", err)
					logrus.Errorf("Order params: %v", params)
					logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
						pp.pair.GetPair(), binance.SideTypeBuy, params.Quantity, params.Price)
					pp.stop <- os.Interrupt
					return
				}
				pp.minuteOrderLimit.Limit++
				pp.dayOrderLimit.Limit++
				if order.Status == binance.OrderStatusTypeNew {
					orderExecutionGuard := pp.OrderExecutionGuard(order)
					<-orderExecutionGuard
					startBuyOrderEvent <- order
				} else {
					for _, fill := range order.Fills {
						fillPrice := utils.ConvStrToFloat64(fill.Price)
						fillQuantity := utils.ConvStrToFloat64(fill.Quantity)
						pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
						pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
						pp.pair.CalcMiddlePrice()
						pp.pair.AddCommission(fill)
					}
					pp.config.Save()
				}
			}
			time.Sleep(pp.pair.GetSleepingTime())
		}
	}()
	return
}

// func ProcessAfterBuyOrder(
// 	config *config_types.ConfigFile,
// 	client *binance.Client,
// 	pair pairs_interfaces.Pairs,
// 	pairInfo *symbol_info_types.SpotSymbol,
// 	minuteOrderLimit *exchange_types.RateLimits,
// 	dayOrderLimit *exchange_types.RateLimits,
// 	minuteRawRequestLimit *exchange_types.RateLimits,
// 	buyEvent chan *pair_price_types.PairPrice,
// 	stopEvent chan os.Signal,
// 	orderStatusEvent chan *binance.WsUserDataEvent,
// 	stopBuy chan bool,
// 	stopSell chan bool,
// 	startBuyOrderEvent chan *binance.CreateOrderResponse) {
// 	go func() {
// 		for {
// 			select {
// 			case <-stopBuy:
// 				stopBuy <- true
// 				return
// 			case <-stopSell:
// 				stopSell <- true
// 				return
// 			case <-stopEvent:
// 				stopEvent <- os.Interrupt
// 				return
// 			case order := <-startBuyOrderEvent:
// 				if order != nil {
// 					for {
// 						orderEvent := <-orderStatusEvent
// 						logrus.Debug("Order status changed")
// 						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
// 							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
// 								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
// 								pair.SetBuyQuantity(pair.GetBuyQuantity() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
// 								pair.SetBuyValue(pair.GetBuyValue() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
// 								pair.CalcMiddlePrice()
// 								config.Save()
// 								break
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}()
// }

// func ProcessAfterSellOrder(
// 	config *config_types.ConfigFile,
// 	client *binance.Client,
// 	pair pairs_interfaces.Pairs,
// 	pairInfo *symbol_info_types.SpotSymbol,
// 	minuteOrderLimit *exchange_types.RateLimits,
// 	dayOrderLimit *exchange_types.RateLimits,
// 	minuteRawRequestLimit *exchange_types.RateLimits,
// 	sellEvent chan *pair_price_types.PairPrice,
// 	stopBuy chan bool,
// 	stopSell chan bool,
// 	stopEvent chan os.Signal,
// 	orderStatusEvent chan *binance.WsUserDataEvent,
// 	startSellOrderEvent chan *binance.CreateOrderResponse) {
// 	go func() {
// 		for {
// 			select {
// 			case <-stopBuy:
// 				stopBuy <- true
// 				return
// 			case <-stopSell:
// 				stopSell <- true
// 				return
// 			case <-stopEvent:
// 				stopEvent <- os.Interrupt
// 				return
// 			case order := <-startSellOrderEvent:
// 				if order != nil {
// 					for {
// 						orderEvent := <-orderStatusEvent
// 						logrus.Debug("Order status changed")
// 						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
// 							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
// 								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
// 								pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
// 								pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
// 								pair.CalcMiddlePrice()
// 								config.Save()
// 								break
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}()
// }

// func OrderExecutionGuard(
// 	config *config_types.ConfigFile,
// 	client *binance.Client,
// 	pair pairs_interfaces.Pairs,
// 	pairInfo *symbol_info_types.SpotSymbol,
// 	minuteOrderLimit *exchange_types.RateLimits,
// 	dayOrderLimit *exchange_types.RateLimits,
// 	minuteRawRequestLimit *exchange_types.RateLimits,
// 	stopProcess chan bool,
// 	stopEvent chan os.Signal,
// 	orderStatusEvent chan *binance.WsUserDataEvent,
// 	order *binance.CreateOrderResponse) (orderExecuted chan bool) {
// 	orderExecuted = make(chan bool)
// 	go func() {
// 		for {
// 			select {
// 			case <-stopProcess:
// 				stopProcess <- true
// 				return
// 			case <-stopEvent:
// 				stopEvent <- os.Interrupt
// 				return
// 			case orderEvent := <-orderStatusEvent:
// 				logrus.Debug("Order status changed")
// 				if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
// 					if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
// 						orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
// 						pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
// 						pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
// 						pair.CalcMiddlePrice()
// 						pair.SetStage(pairs_types.PositionClosedStage)
// 						config.Save()
// 						orderExecuted <- true
// 						return
// 					}
// 				}
// 			}
// 		}
// 	}()
// 	return
// }

func (pp *PairProcessor) LimitUpdaterStream() {

	go func() {
		for {
			select {
			case <-time.After(pp.updateTime):
				pp.updateTime,
					pp.minuteOrderLimit,
					pp.dayOrderLimit,
					pp.minuteRawRequestLimit = LimitRead(pp.degree, []string{pp.pair.GetPair()}, pp.client)
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			}
		}
	}()

	// Перевіряємо чи не вийшли за ліміти на запити та ордери
	go func() {
		for {
			select {
			case <-pp.stop:
				pp.stop <- os.Interrupt
			case <-pp.limitsOut:
				pp.stop <- os.Interrupt
				return
			default:
			}
			time.Sleep(pp.updateTime)
		}
	}()
}

func (pp *PairProcessor) OrderExecutionGuard(
	// config *config_types.ConfigFile,
	// client *binance.Client,
	// pair pairs_interfaces.Pairs,
	// pairInfo *symbol_info_types.SpotSymbol,
	// minuteOrderLimit *exchange_types.RateLimits,
	// dayOrderLimit *exchange_types.RateLimits,
	// minuteRawRequestLimit *exchange_types.RateLimits,
	// stopProcess chan bool,
	// stopEvent chan os.Signal,
	// orderStatusEvent chan *binance.WsUserDataEvent,
	order *binance.CreateOrderResponse) (orderExecuted chan bool) {
	orderExecuted = pp.orderExecuted
	go func() {
		for {
			select {
			case <-pp.stopAfterProcess:
				pp.stopAfterProcess <- true
				return
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case orderEvent := <-pp.orderStatusEvent:
				logrus.Debug("Order status changed")
				if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
					if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
						orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
						pp.pair.SetSellQuantity(pp.pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
						pp.pair.SetSellValue(pp.pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
						pp.pair.CalcMiddlePrice()
						pp.pair.SetStage(pairs_types.PositionClosedStage)
						pp.config.Save()
						pp.orderExecuted <- true
						return
					}
				}
			}
		}
	}()
	return
}

func NewPairProcessor(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	orderType binance.OrderType,
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) (pp *PairProcessor, err error) {
	pp = &PairProcessor{
		client:                client,
		pair:                  pair,
		account:               nil,
		stop:                  make(chan os.Signal, 1),
		limitsOut:             make(chan bool, 1),
		pairInfo:              nil,
		orderType:             orderType,
		buyEvent:              buyEvent,
		sellEvent:             sellEvent,
		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},
		stopBuy:               make(chan bool, 1),
		stopSell:              make(chan bool, 1),
		stopAfterProcess:      make(chan bool, 1),
		orderExecuted:         make(chan bool, 1),
		orderStatusEvent:      nil,
		degree:                3,
	}
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit =
		LimitRead(degree, []string{pp.pair.GetPair()}, client)

	pp.exchangeInfo = exchange_types.New()
	err = spot_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}

	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.SpotSymbol{Symbol: pair.GetPair()}).(*symbol_types.SpotSymbol)

	pp.LimitUpdaterStream()

	listenKey, err := pp.client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return
	}
	pp.userDataStream4Order = spot_streams.NewUserDataStream(listenKey, 1)
	pp.userDataStream4Order.Start()

	orderStatuses := []binance.OrderStatusType{
		binance.OrderStatusTypeFilled,
		binance.OrderStatusTypePartiallyFilled,
	}
	pp.orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(
		pp.userDataStream4Order.GetDataChannel(),
		orderStatuses)

	return
}
