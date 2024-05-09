package spot_signals

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairProcessor struct {
		config                               *config_types.ConfigFile
		client                               *binance.Client
		pair                                 pairs_interfaces.Pairs
		exchangeInfo                         *exchange_types.ExchangeInfo
		account                              *spot_account.Account
		orderType                            binance.OrderType
		buyEvent                             chan *pair_price_types.PairPrice
		sellEvent                            chan *pair_price_types.PairPrice
		startBuyOrderEvent                   chan *binance.CreateOrderResponse
		startSellOrderEvent                  chan *binance.CreateOrderResponse
		startProcessBuyTakeProfitOrderEvent  chan *binance.CreateOrderResponse
		startProcessSellTakeProfitOrderEvent chan *binance.CreateOrderResponse
		updateTime                           time.Duration
		minuteOrderLimit                     *exchange_types.RateLimits
		dayOrderLimit                        *exchange_types.RateLimits
		minuteRawRequestLimit                *exchange_types.RateLimits
		stop                                 chan os.Signal
		limitsOut                            chan bool
		stopBuy                              chan bool
		stopSell                             chan bool
		stopAfterProcess                     chan bool
		orderExecuted                        chan bool
		userDataEvent                        chan *binance.WsUserDataEvent
		orderStatusEvent                     chan *binance.WsUserDataEvent
		pairInfo                             *symbol_types.SpotSymbol
		degree                               int
		debug                                bool
	}
)

func (pp *PairProcessor) ProcessBuyOrder() (startBuyOrderEvent chan *binance.CreateOrderResponse, err error) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
	if err != nil {
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	if pp.buyEvent == nil && pp.sellEvent == nil && pp.startBuyOrderEvent == nil {
		pp.buyEvent = make(chan *pair_price_types.PairPrice)
		pp.sellEvent = make(chan *pair_price_types.PairPrice)
		pp.startBuyOrderEvent = make(chan *binance.CreateOrderResponse)
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
					if !pp.debug {
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
							pp.startBuyOrderEvent <- order
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
					} else {
						pp.pair.SetBuyQuantity(params.Quantity)
						pp.pair.SetBuyValue(params.Quantity * params.Price)
						pp.pair.CalcMiddlePrice()
						pp.config.Save()
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	} else {
		startBuyOrderEvent = pp.startBuyOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellOrder() (startSellOrderEvent chan *binance.CreateOrderResponse, err error) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
	if err != nil {
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)

	if pp.buyEvent == nil && pp.sellEvent == nil && pp.startSellOrderEvent == nil {
		pp.buyEvent = make(chan *pair_price_types.PairPrice)
		pp.sellEvent = make(chan *pair_price_types.PairPrice)
		pp.startSellOrderEvent = make(chan *binance.CreateOrderResponse)
		startSellOrderEvent = pp.startSellOrderEvent
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
					if !pp.debug {
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
					} else {
						pp.pair.SetBuyQuantity(params.Quantity)
						pp.pair.SetBuyValue(params.Quantity * params.Price)
						pp.pair.CalcMiddlePrice()
						pp.config.Save()
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	} else {
		startSellOrderEvent = pp.startSellOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessBuyTakeProfitOrder() (startProcessBuyTakeProfitOrderEvent chan *binance.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	if pp.startProcessBuyTakeProfitOrderEvent == nil {
		pp.startProcessBuyTakeProfitOrderEvent = make(chan *binance.CreateOrderResponse)
		startProcessBuyTakeProfitOrderEvent = pp.startProcessBuyTakeProfitOrderEvent
		go func() {
			for {
				select {
				case <-pp.stopAfterProcess:
					pp.stopAfterProcess <- true
					return
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case params := <-pp.buyEvent:
					if pp.minuteOrderLimit.Limit == 0 || pp.dayOrderLimit.Limit == 0 || pp.minuteRawRequestLimit.Limit == 0 {
						logrus.Warn("Order limits has been out!!!, waiting for update...")
						continue
					}
					if !pp.debug {
						order, err :=
							pp.client.NewCreateOrderService().
								Symbol(string(binance.SymbolType(pp.pair.GetPair()))).
								Type(binance.OrderTypeTakeProfit).
								Side(binance.SideTypeBuy).
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
							startProcessBuyTakeProfitOrderEvent <- order
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
					} else {
						pp.pair.SetBuyQuantity(params.Quantity)
						pp.pair.SetBuyValue(params.Quantity * params.Price)
						pp.pair.CalcMiddlePrice()
						pp.config.Save()
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	} else {
		startProcessBuyTakeProfitOrderEvent = pp.startProcessBuyTakeProfitOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellTakeProfitOrder() (startProcessSellTakeProfitOrderEvent chan *binance.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	if pp.startProcessSellTakeProfitOrderEvent == nil {
		pp.startProcessSellTakeProfitOrderEvent = make(chan *binance.CreateOrderResponse)
		startProcessSellTakeProfitOrderEvent = pp.startProcessSellTakeProfitOrderEvent
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
					if !pp.debug {
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
								pp.pair.GetPair(), binance.SideTypeSell, params.Quantity, params.Price)
							pp.stop <- os.Interrupt
							return
						}
						pp.minuteOrderLimit.Limit++
						pp.dayOrderLimit.Limit++
						if order.Status == binance.OrderStatusTypeNew {
							orderExecutionGuard := pp.OrderExecutionGuard(order)
							<-orderExecutionGuard
							startProcessSellTakeProfitOrderEvent <- order
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
					} else {
						pp.pair.SetBuyQuantity(params.Quantity)
						pp.pair.SetBuyValue(params.Quantity * params.Price)
						pp.pair.CalcMiddlePrice()
						pp.config.Save()
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	} else {
		startProcessSellTakeProfitOrderEvent = pp.startProcessSellTakeProfitOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessAfterBuyOrder(startBuyOrderEvent chan *binance.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-pp.stopBuy:
				pp.stopBuy <- true
				return
			case <-pp.stopSell:
				pp.stopSell <- true
				return
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case order := <-startBuyOrderEvent:
				if order != nil {
					for {
						orderEvent := <-pp.orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
								pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
								pp.pair.SetBuyValue(pp.pair.GetBuyValue() - utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
								pp.pair.CalcMiddlePrice()
								pp.config.Save()
								break
							}
						}
					}
				}
			}
		}
	}()
}

func (pp *PairProcessor) ProcessAfterSellOrder(startPostProcessOrderEvent chan *binance.CreateOrderResponse) {
	go func() {
		for {
			select {
			case <-pp.stopBuy:
				pp.stopBuy <- true
				return
			case <-pp.stopSell:
				pp.stopSell <- true
				return
			case <-pp.stop:
				pp.stop <- os.Interrupt
				return
			case order := <-startPostProcessOrderEvent:
				if order != nil {
					for {
						orderEvent := <-pp.orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderUpdate.Id == order.OrderID || orderEvent.OrderUpdate.ClientOrderId == order.ClientOrderID {
							if orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) ||
								orderEvent.OrderUpdate.Status == string(binance.OrderStatusTypePartiallyFilled) {
								pp.pair.SetSellQuantity(pp.pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume))
								pp.pair.SetSellValue(pp.pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderUpdate.Volume)*utils.ConvStrToFloat64(orderEvent.OrderUpdate.Price))
								pp.pair.CalcMiddlePrice()
								pp.config.Save()
								break
							}
						}
					}
				}
			}
		}
	}()
}

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

func (pp *PairProcessor) OrderExecutionGuard(order *binance.CreateOrderResponse) chan bool {
	if pp.orderExecuted == nil {
		pp.orderExecuted = make(chan bool, 1)
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
	}
	return pp.orderExecuted
}

func NewPairProcessor(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	orderType binance.OrderType,
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice,
	debug bool) (pp *PairProcessor, err error) {
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
		orderExecuted:         nil,
		orderStatusEvent:      nil,
		degree:                3,
		debug:                 debug,
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
	pp.userDataEvent = make(chan *binance.WsUserDataEvent)
	_, _, err = binance.WsUserDataServe(listenKey, func(event *binance.WsUserDataEvent) {
		pp.userDataEvent <- event
	}, utils.HandleErr)
	if err != nil {
		return
	}

	orderStatuses := []binance.OrderStatusType{
		binance.OrderStatusTypeFilled,
		binance.OrderStatusTypePartiallyFilled,
	}
	pp.orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(pp.userDataEvent, orderStatuses)

	return
}
