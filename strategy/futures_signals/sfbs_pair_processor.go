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

func (pp *PairProcessor) StopBuySignal() {
	pp.stopBuy <- true
}

func (pp *PairProcessor) StopSellSignal() {
	pp.stopSell <- true
}

//  1. Order with type STOP, parameter timeInForce can be sent ( default GTC).
//  2. Order with type TAKE_PROFIT, parameter timeInForce can be sent ( default GTC).
//  3. Condition orders will be triggered when:
//     a) If parameterpriceProtectis sent as true:
//     when price reaches the stopPrice ，the difference rate between "MARK_PRICE" and
//     "CONTRACT_PRICE" cannot be larger than the "triggerProtect" of the symbol
//     "triggerProtect" of a symbol can be got from GET /fapi/v1/exchangeInfo
//     b) STOP, STOP_MARKET:
//     BUY: latest price ("MARK_PRICE" or "CONTRACT_PRICE") >= stopPrice
//     SELL: latest price ("MARK_PRICE" or "CONTRACT_PRICE") <= stopPrice
//     c) TAKE_PROFIT, TAKE_PROFIT_MARKET:
//     BUY: latest price ("MARK_PRICE" or "CONTRACT_PRICE") <= stopPrice
//     SELL: latest price ("MARK_PRICE" or "CONTRACT_PRICE") >= stopPrice
//     d) TRAILING_STOP_MARKET:
//     BUY: the lowest price after order placed <= activationPrice,
//     and the latest price >= the lowest price * (1 + callbackRate)
//     SELL: the highest price after order placed >= activationPrice,
//     and the latest price <= the highest price * (1 - callbackRate)
//  4. For TRAILING_STOP_MARKET, if you got such error code.
//     {"code": -2021, "msg": "Order would immediately trigger."}
//     means that the parameters you send do not meet the following requirements:
//     BUY: activationPrice should be smaller than latest price.
//     SELL: activationPrice should be larger than latest price.
//     If newOrderRespType is sent as RESULT :
//     MARKET order: the final FILLED result of the order will be return directly.
//     LIMIT order with special timeInForce:
//     the final status result of the order(FILLED or EXPIRED)
//     will be returned directly.
//  5. STOP_MARKET, TAKE_PROFIT_MARKET with closePosition=true:
//     Follow the same rules for condition orders.
//     If triggered，close all current long position( if SELL) or current short position( if BUY).
//     Cannot be used with quantity parameter
//     Cannot be used with reduceOnly parameter
//     In Hedge Mode,cannot be used with BUY orders in LONG position side
//     and cannot be used with SELL orders in SHORT position side
//  6. selfTradePreventionMode is only effective when timeInForce set to IOC or GTC or GTD.
//  7. In extreme market conditions,
//     timeInForce GTD order auto cancel time might be delayed comparing to goodTillDate
func (pp *PairProcessor) CreateOrder(
	orderType futures.OrderType,
	sideType futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity float64,
	closePosition bool,
	price float64,
	stopPrice float64,
	callbackRate float64) (
	order *futures.CreateOrderResponse, err error) {
	symbol, err := (*pp.pairInfo).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
			Type(orderType).
			Side(sideType)
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == futures.OrderTypeMarket {
		// MARKET	quantity
		service = service.Quantity(utils.ConvFloat64ToStr(quantity, quantityRound))
	} else if orderType == futures.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
	} else if orderType == futures.OrderTypeStop || orderType == futures.OrderTypeTakeProfit {
		// STOP/TAKE_PROFIT	quantity, price, stopPrice
		service = service.
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound)).
			StopPrice(utils.ConvFloat64ToStr(stopPrice, priceRound))
	} else if orderType == futures.OrderTypeStopMarket || orderType == futures.OrderTypeTakeProfitMarket {
		// STOP_MARKET/TAKE_PROFIT_MARKET	stopPrice
		service = service.
			StopPrice(utils.ConvFloat64ToStr(stopPrice, priceRound))
		if closePosition {
			service = service.ClosePosition(closePosition)
		}
	} else if orderType == futures.OrderTypeTrailingStopMarket {
		// TRAILING_STOP_MARKET	quantity,callbackRate
		service = service.
			TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			CallbackRate(utils.ConvFloat64ToStr(callbackRate, priceRound))
		if stopPrice != 0 {
			service = service.
				ActivationPrice(utils.ConvFloat64ToStr(stopPrice, priceRound))
		}
	}
	return service.Do(context.Background())
}

func (pp *PairProcessor) ProcessBuyOrder() (startBuyOrderEvent chan *futures.CreateOrderResponse) {
	go func() {
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
					order, err := pp.CreateOrder(
						pp.orderType,
						futures.SideTypeBuy,
						futures.TimeInForceTypeGTC,
						params.Quantity,
						false, // We close position manually
						params.Price,
						0,
						0)
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
	startSellOrderEvent = make(chan *futures.CreateOrderResponse)
	go func() {
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
					order, err := pp.CreateOrder(
						pp.orderType,
						futures.SideTypeBuy,
						futures.TimeInForceTypeGTC,
						params.Quantity,
						false, // We close position manually
						params.Price,
						0,
						0)
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

func (pp *PairProcessor) ProcessBuyTakeProfitOrder(trailingDelta int) (startProcessBuyTakeProfitEvent chan *futures.CreateOrderResponse) {
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
						order, err := pp.CreateOrder(
							futures.OrderTypeTakeProfit,
							futures.SideTypeBuy,
							futures.TimeInForceTypeGTC,
							params.Quantity,
							false, // We close position manually
							params.Price,
							params.Price*(1-float64(trailingDelta)/100),
							0)
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

func (pp *PairProcessor) ProcessSellTakeProfitOrder(trailingDelta int) (startProcessSellTakeProfitEvent chan *futures.CreateOrderResponse) {
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
						order, err := pp.CreateOrder(
							futures.OrderTypeTakeProfit,
							futures.SideTypeSell,
							futures.TimeInForceTypeGTC,
							params.Quantity,
							false, // We close position manually
							params.Price,
							params.Price*(1+float64(trailingDelta)/100),
							0)
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
