package futures_signals

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"

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
		config                          *config_types.ConfigFile
		client                          *futures.Client
		pair                            pairs_interfaces.Pairs
		exchangeInfo                    *exchange_types.ExchangeInfo
		account                         *futures_account.Account
		orderType                       futures.OrderType
		buyEvent                        chan *pair_price_types.PairPrice
		sellEvent                       chan *pair_price_types.PairPrice
		updateTime                      time.Duration
		minuteOrderLimit                *exchange_types.RateLimits
		dayOrderLimit                   *exchange_types.RateLimits
		minuteRawRequestLimit           *exchange_types.RateLimits
		stop                            chan os.Signal
		limitsOut                       chan bool
		stopBuy                         chan bool
		stopSell                        chan bool
		askUp                           chan *pair_price_types.AskBid
		askDown                         chan *pair_price_types.AskBid
		bidUp                           chan *pair_price_types.AskBid
		bidDown                         chan *pair_price_types.AskBid
		stopAfterProcess                chan bool
		orderExecuted                   chan bool
		userDataEvent                   chan *futures.WsUserDataEvent
		orderStatusEvent                chan *futures.WsUserDataEvent
		pairInfo                        *symbol_types.FuturesSymbol
		degree                          int
		debug                           bool
		startProcessSellTakeProfitEvent chan *futures.CreateOrderResponse
		startProcessBuyTakeProfitEvent  chan *futures.CreateOrderResponse
	}
)

func (pp *PairProcessor) ProcessBuyOrder() (startBuyOrderEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetFuturesSymbol()
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
				if !pp.debug {
					service :=
						pp.client.NewCreateOrderService().
							Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
							Type(pp.orderType).
							Side(futures.SideTypeBuy).
							Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
					if pp.orderType == futures.OrderTypeMarket {
						order, err = service.Do(context.Background())
					} else if pp.orderType == futures.OrderTypeLimit {
						order, err = service.
							Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
							TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
					}
					if err != nil {
						logrus.Errorf("Can't create order: %v", err)
						logrus.Errorf("Order params: %v", params)
						logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
							pp.pair.GetPair(), futures.SideTypeBuy, params.Quantity, params.Price)
						pp.stop <- os.Interrupt
						return
					}
					pp.minuteOrderLimit.Limit++
					pp.dayOrderLimit.Limit++
					if order.Status == futures.OrderStatusTypeNew {
						startBuyOrderEvent <- order
					} else {
						fillPrice := utils.ConvStrToFloat64(order.Price)
						fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
						pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
						pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
						pp.pair.CalcMiddlePrice()
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
	return
}

func (pp *PairProcessor) ProcessSellOrder() (startSellOrderEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetFuturesSymbol()
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
							Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
							Type(futures.OrderTypeLimit).
							Side(futures.SideTypeSell).
							Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound))
					if pp.orderType == futures.OrderTypeMarket {
						order, err = service.Do(context.Background())
					} else if pp.orderType == futures.OrderTypeLimit {
						order, err = service.
							Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
							TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
					}
					if err != nil {
						logrus.Errorf("Can't create order: %v", err)
						logrus.Errorf("Order params: %v", params)
						logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
							pp.pair.GetPair(), futures.SideTypeSell, params.Quantity, params.Price)
						pp.stop <- os.Interrupt
						return
					}
					pp.minuteOrderLimit.Limit++
					pp.dayOrderLimit.Limit++
					if order.Status == futures.OrderStatusTypeNew {
						startSellOrderEvent <- order
					} else {
						fillPrice := utils.ConvStrToFloat64(order.Price)
						fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
						pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
						pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
						pp.pair.CalcMiddlePrice()
						pp.config.Save()
					}
				} else {
					pp.pair.SetSellQuantity(params.Quantity)
					pp.pair.SetSellValue(params.Quantity * params.Price)
					pp.pair.CalcMiddlePrice()
					pp.config.Save()
				}
			}
			time.Sleep(pp.pair.GetSleepingTime())
		}
	}()
	return
}

func (pp *PairProcessor) ProcessBuyTakeProfitOrder() (startProcessBuyTakeProfitEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	if pp.startProcessBuyTakeProfitEvent == nil {
		pp.startProcessBuyTakeProfitEvent = make(chan *futures.CreateOrderResponse)
		startProcessBuyTakeProfitEvent = pp.startProcessBuyTakeProfitEvent
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
								Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
								Type(futures.OrderTypeTakeProfit).
								Side(futures.SideTypeBuy).
								Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
								Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
								TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
						if err != nil {
							logrus.Errorf("Can't create order: %v", err)
							logrus.Errorf("Order params: %v", params)
							logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
								pp.pair.GetPair(), futures.SideTypeBuy, params.Quantity, params.Price)
							pp.stop <- os.Interrupt
							return
						}
						pp.minuteOrderLimit.Limit++
						pp.dayOrderLimit.Limit++
						if order.Status == futures.OrderStatusTypeNew {
							orderExecutionGuard := pp.OrderExecutionGuard(order)
							<-orderExecutionGuard
							startProcessBuyTakeProfitEvent <- order
						} else {
							fillPrice := utils.ConvStrToFloat64(order.Price)
							fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
							pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
							pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
							pp.pair.CalcMiddlePrice()
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
	}
	return
}

func (pp *PairProcessor) ProcessSellTakeProfitOrder() (startProcessSellTakeProfitEvent chan *futures.CreateOrderResponse) {
	symbol, err := (*pp.pairInfo).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	if pp.startProcessSellTakeProfitEvent == nil {
		pp.startProcessSellTakeProfitEvent = make(chan *futures.CreateOrderResponse)
		startProcessSellTakeProfitEvent = pp.startProcessSellTakeProfitEvent
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
								Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
								Type(futures.OrderTypeTakeProfit).
								Side(futures.SideTypeSell).
								Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
								Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
								TimeInForce(futures.TimeInForceTypeGTC).Do(context.Background())
						if err != nil {
							logrus.Errorf("Can't create order: %v", err)
							logrus.Errorf("Order params: %v", params)
							logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
								pp.pair.GetPair(), futures.SideTypeSell, params.Quantity, params.Price)
							pp.stop <- os.Interrupt
							return
						}
						pp.minuteOrderLimit.Limit++
						pp.dayOrderLimit.Limit++
						if order.Status == futures.OrderStatusTypeNew {
							orderExecutionGuard := pp.OrderExecutionGuard(order)
							<-orderExecutionGuard
							startProcessSellTakeProfitEvent <- order
						} else {
							fillPrice := utils.ConvStrToFloat64(order.Price)
							fillQuantity := utils.ConvStrToFloat64(order.ExecutedQuantity)
							pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() + fillQuantity)
							pp.pair.SetBuyValue(pp.pair.GetBuyValue() + fillQuantity*fillPrice)
							pp.pair.CalcMiddlePrice()
							pp.config.Save()
						}
					} else {
						pp.pair.SetSellQuantity(params.Quantity)
						pp.pair.SetSellValue(params.Quantity * params.Price)
						pp.pair.CalcMiddlePrice()
						pp.config.Save()
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return
}

func (pp *PairProcessor) ProcessAfterBuyOrder(startProcessAfterBuyOrderEvent chan *futures.CreateOrderResponse) {
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
			case order := <-startProcessAfterBuyOrderEvent:
				if order != nil {
					for {
						orderEvent := <-pp.orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderTradeUpdate.ID == order.OrderID || orderEvent.OrderTradeUpdate.ClientOrderID == order.ClientOrderID {
							if orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
								orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled {
								pp.pair.SetBuyQuantity(pp.pair.GetBuyQuantity() - utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty))
								pp.pair.SetBuyValue(pp.pair.GetBuyValue() - utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty)*utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledPrice))
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

func (pp *PairProcessor) ProcessAfterSellOrder(startSellOrderEvent chan *futures.CreateOrderResponse) {
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
			case order := <-startSellOrderEvent:
				if order != nil {
					for {
						orderEvent := <-pp.orderStatusEvent
						logrus.Debug("Order status changed")
						if orderEvent.OrderTradeUpdate.ID == order.OrderID || orderEvent.OrderTradeUpdate.ClientOrderID == order.ClientOrderID {
							if orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
								orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled {
								pp.pair.SetSellQuantity(pp.pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty))
								pp.pair.SetSellValue(pp.pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty)*utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledPrice))
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

func (pp *PairProcessor) OrderExecutionGuard(order *futures.CreateOrderResponse) chan bool {
	if pp.orderExecuted == nil {
		pp.orderExecuted = make(chan bool)
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
					if orderEvent.OrderTradeUpdate.ID == order.OrderID || orderEvent.OrderTradeUpdate.ClientOrderID == order.ClientOrderID {
						if orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
							orderEvent.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled {
							pp.pair.SetSellQuantity(pp.pair.GetSellQuantity() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty))
							pp.pair.SetSellValue(pp.pair.GetSellValue() + utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledQty)*utils.ConvStrToFloat64(orderEvent.OrderTradeUpdate.LastFilledPrice))
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
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	orderType futures.OrderType,
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice,
	askUp chan *pair_price_types.AskBid,
	askDown chan *pair_price_types.AskBid,
	bidUp chan *pair_price_types.AskBid,
	bidDown chan *pair_price_types.AskBid,
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
		askUp:                 askUp,
		askDown:               askDown,
		bidUp:                 bidUp,
		bidDown:               bidDown,
		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},
		stopBuy:               nil,
		stopSell:              nil,
		stopAfterProcess:      nil,
		orderExecuted:         nil,
		orderStatusEvent:      nil,
		degree:                3,
		debug:                 debug,
	}
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit =
		LimitRead(pp.degree, []string{pp.pair.GetPair()}, client)

	pp.exchangeInfo = exchange_types.New()
	err = futures_exchange_info.Init(pp.exchangeInfo, pp.degree, client)
	if err != nil {
		return
	}

	pp.account, err = futures_account.New(pp.client, pp.degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: pair.GetPair()}).(*symbol_types.FuturesSymbol)

	pp.LimitUpdaterStream()

	listenKey, err := pp.client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return
	}
	pp.userDataEvent = make(chan *futures.WsUserDataEvent)
	_, _, err = futures.WsUserDataServe(listenKey, func(event *futures.WsUserDataEvent) {
		pp.userDataEvent <- event
	}, utils.HandleErr)
	if err != nil {
		return
	}

	orderStatuses := []futures.OrderStatusType{
		futures.OrderStatusTypeFilled,
		futures.OrderStatusTypePartiallyFilled,
	}
	pp.orderStatusEvent = futures_handlers.GetChangingOfOrdersGuard(pp.userDataEvent, orderStatuses)

	return
}
