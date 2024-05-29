package futures_signals

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"

	utils "github.com/fr0ster/go-trading-utils/utils"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	PairProcessor struct {
		config       *config_types.ConfigFile
		client       *futures.Client
		pair         *pairs_types.Pairs
		exchangeInfo *exchange_types.ExchangeInfo
		account      *futures_account.Account

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

		buyEvent       chan *pair_price_types.PairPrice
		stopBuy        chan bool
		buyProcessRun  bool
		buyOrderEvent  chan *futures.CreateOrderResponse
		sellEvent      chan *pair_price_types.PairPrice
		stopSell       chan bool
		sellProcessRun bool
		sellOrderEvent chan *futures.CreateOrderResponse

		startProcessBuyTakeProfitEvent  chan *futures.CreateOrderResponse
		buyTakeProfitProcessRun         bool
		stopBuyTakeProfitProcess        chan bool
		startProcessSellTakeProfitEvent chan *futures.CreateOrderResponse
		sellTakeProfitProcessRun        bool
		stopSellTakeProfitProcess       chan bool

		orderExecuted                  chan bool
		orderExecutionGuardProcessRun  bool
		stopOrderExecutionGuardProcess chan bool

		userDataEvent    chan *futures.WsUserDataEvent
		orderStatusEvent chan *futures.WsUserDataEvent

		stop      chan os.Signal
		limitsOut chan bool

		pairInfo     *symbol_types.FuturesSymbol
		orderTypes   map[futures.OrderType]bool
		degree       int
		debug        bool
		sleepingTime time.Duration
		timeOut      time.Duration
	}
)

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
	if _, ok := pp.orderTypes[orderType]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.pair.GetPair())
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderResponseType(futures.NewOrderRespTypeRESULT).
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

func (pp *PairProcessor) ClosePosition() (res *futures.CreateOrderResponse, err error) {
	var (
		side = futures.SideTypeBuy
	)
	risk, err := pp.account.GetPositionRisk(pp.pair.GetPair())
	if err != nil {
		logrus.Errorf(errorMsg, err)
		return
	}
	if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		side = futures.SideTypeSell
	} else if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		side = futures.SideTypeBuy
	} else {
		return
	}
	return pp.client.NewCreateOrderService().
		Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
		Type(futures.OrderTypeMarket).
		Side(side).
		StopPrice(risk.EntryPrice).
		ClosePosition(true).
		Do(context.Background())
}

func (pp *PairProcessor) ProcessBuyOrder(triggerEvent chan *pair_price_types.PairPrice) (nextTriggerEvent chan *futures.CreateOrderResponse) {
	if !pp.buyProcessRun {
		if pp.buyEvent == nil {
			pp.buyEvent = triggerEvent
		}
		if pp.buyOrderEvent == nil {
			pp.buyOrderEvent = make(chan *futures.CreateOrderResponse)
		}
		nextTriggerEvent = pp.buyOrderEvent
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
							futures.OrderTypeMarket,
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
							nextTriggerEvent <- order
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
				time.Sleep(pp.sleepingTime)
			}
		}()
	} else {
		nextTriggerEvent = pp.buyOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellOrder(triggerEvent chan *pair_price_types.PairPrice) (startSellOrderEvent chan *futures.CreateOrderResponse) {
	if !pp.sellProcessRun {
		if pp.sellEvent == nil {
			pp.sellEvent = triggerEvent
		}
		if pp.sellOrderEvent == nil {
			pp.sellOrderEvent = make(chan *futures.CreateOrderResponse, 1)
		}
		startSellOrderEvent = make(chan *futures.CreateOrderResponse, 1)
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
							futures.OrderTypeMarket,
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
				time.Sleep(pp.sleepingTime)
			}
		}()
		pp.sellProcessRun = true
	} else {
		startSellOrderEvent = pp.sellOrderEvent
	}
	return
}

func (pp *PairProcessor) StopBuySignal() {
	if pp.buyProcessRun {
		pp.buyProcessRun = false
		pp.stopBuy <- true
	}
}

func (pp *PairProcessor) StopSellSignal() {
	if pp.sellProcessRun {
		pp.sellProcessRun = false
		pp.stopSell <- true
	}
}

func (pp *PairProcessor) ProcessBuyTakeProfitOrder(trailingDelta int) (startProcessBuyTakeProfitEvent chan *futures.CreateOrderResponse) {
	if !pp.buyTakeProfitProcessRun {
		if pp.startProcessBuyTakeProfitEvent == nil {
			pp.startProcessBuyTakeProfitEvent = make(chan *futures.CreateOrderResponse)
		}
		startProcessBuyTakeProfitEvent = pp.startProcessBuyTakeProfitEvent
		go func() {
			for {
				select {
				case <-pp.stopBuyTakeProfitProcess:
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
				time.Sleep(pp.sleepingTime)
			}
		}()
		pp.buyTakeProfitProcessRun = true
	} else {
		startProcessBuyTakeProfitEvent = pp.startProcessBuyTakeProfitEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellTakeProfitOrder(trailingDelta int) (startProcessSellTakeProfitEvent chan *futures.CreateOrderResponse) {
	if !pp.sellTakeProfitProcessRun {
		if pp.startProcessSellTakeProfitEvent == nil {
			pp.startProcessSellTakeProfitEvent = make(chan *futures.CreateOrderResponse)
		}
		startProcessSellTakeProfitEvent = pp.startProcessSellTakeProfitEvent
		go func() {
			for {
				select {
				case <-pp.stopSellTakeProfitProcess:
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
				time.Sleep(pp.sleepingTime)
			}
		}()
		pp.sellTakeProfitProcessRun = true
	} else {
		startProcessSellTakeProfitEvent = pp.startProcessSellTakeProfitEvent
	}
	return
}

func (pp *PairProcessor) StopBuyTakeProfitSignal() {
	if pp.buyTakeProfitProcessRun {
		pp.buyTakeProfitProcessRun = false
		pp.stopBuyTakeProfitProcess <- true
	}
}

func (pp *PairProcessor) StopSellTakeProfitSignal() {
	if pp.sellTakeProfitProcessRun {
		pp.sellTakeProfitProcessRun = false
		pp.stopSellTakeProfitProcess <- true
	}
}

func (pp *PairProcessor) ProcessAfterBuyOrder(triggerEvent chan *futures.CreateOrderResponse) {
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
			case order := <-triggerEvent:
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

func (pp *PairProcessor) ProcessAfterSellOrder(triggerEvent chan *futures.CreateOrderResponse) {
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
			case order := <-triggerEvent:
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
	if !pp.orderExecutionGuardProcessRun {
		if pp.orderExecuted == nil {
			pp.orderExecuted = make(chan bool)
		}
		go func() {
			for {
				select {
				case <-pp.stopOrderExecutionGuardProcess:
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
		pp.orderExecutionGuardProcessRun = true
	}
	return pp.orderExecuted
}

func (pp *PairProcessor) StopOrderExecutionGuard() {
	if pp.orderExecutionGuardProcessRun {
		pp.orderExecutionGuardProcessRun = false
		pp.stopOrderExecutionGuardProcess <- true
	}
}

func (pp *PairProcessor) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairProcessor) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func (pp *PairProcessor) CheckOrderType(orderType futures.OrderType) bool {
	_, ok := pp.orderTypes[orderType]
	return ok
}

func (pp *PairProcessor) GetOpenOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *futures.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.pair.GetPair()).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *futures.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.pair.GetPair()).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (err error) {
	return pp.client.NewCancelAllOpenOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetUserDataEvent() chan *futures.WsUserDataEvent {
	return pp.userDataEvent
}

func (pp *PairProcessor) GetOrderStatusEvent() chan *futures.WsUserDataEvent {
	return pp.orderStatusEvent
}

func (pp *PairProcessor) GetPositionRisk() (risks *futures.PositionRisk, err error) {
	risks, err = pp.account.GetPositionRisk(pp.pair.GetPair())
	return
}

func (pp *PairProcessor) GetLeverage() int {
	risk, _ := pp.GetPositionRisk()
	leverage, _ := strconv.Atoi(risk.Leverage) // Convert string to int
	return leverage
}

func (pp *PairProcessor) SetLeverage(leverage int) (res *futures.SymbolLeverage, err error) {
	return pp.client.NewChangeLeverageService().Symbol(pp.pair.GetPair()).Leverage(leverage).Do(context.Background())
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) GetMarginType() pairs_types.MarginType {
	risk, _ := pp.GetPositionRisk()
	return pairs_types.MarginType(strings.ToUpper(risk.MarginType))
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) SetMarginType(marginType pairs_types.MarginType) (err error) {
	return pp.client.
		NewChangeMarginTypeService().
		Symbol(pp.pair.GetPair()).
		MarginType(futures.MarginType(marginType)).
		Do(context.Background())
}

func (pp *PairProcessor) GetPositionMargin() (margin float64) {
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	margin = utils.ConvStrToFloat64(risk.IsolatedMargin) // Convert string to float64
	return
}

func (pp *PairProcessor) SetPositionMargin(amountMargin float64, typeMargin int) (err error) {
	return pp.client.NewUpdatePositionMarginService().
		Symbol(pp.pair.GetPair()).Type(typeMargin).
		Amount(utils.ConvFloat64ToStrDefault(amountMargin)).Do(context.Background())
}

func NewPairProcessor(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	exchangeInfo *exchange_types.ExchangeInfo,
	account *futures_account.Account,
	userDataEvent chan *futures.WsUserDataEvent,
	debug bool) (pp *PairProcessor, err error) {
	pp = &PairProcessor{
		client:       client,
		pair:         pair,
		exchangeInfo: exchangeInfo,
		account:      account,

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		buyEvent:       nil,
		stopBuy:        nil,
		buyProcessRun:  false,
		sellEvent:      nil,
		stopSell:       nil,
		sellProcessRun: false,

		orderExecuted:                  nil,
		orderExecutionGuardProcessRun:  false,
		stopOrderExecutionGuardProcess: nil,

		userDataEvent:    userDataEvent,
		orderStatusEvent: nil,

		stop:      make(chan os.Signal, 1),
		limitsOut: make(chan bool, 1),

		pairInfo:     nil,
		orderTypes:   nil,
		degree:       3,
		debug:        debug,
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}
	// Перевіряємо ліміти на ордери та запити
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit =
		LimitRead(pp.degree, []string{pp.pair.GetPair()}, client)

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: pair.GetPair()}).(*symbol_types.FuturesSymbol)

	// Ініціалізуємо типи ордерів
	pp.orderTypes = make(map[futures.OrderType]bool, 0)
	for _, orderType := range pp.pairInfo.OrderType {
		pp.orderTypes[orderType] = true
	}

	// Визначаємо статуси ордерів які нас цікавлять та ...
	// ... запускаємо стрім для відслідковування зміни статусу ордерів які нас цікавлять
	pp.orderStatusEvent = futures_handlers.GetChangingOfOrdersGuard(pp.userDataEvent, futures.OrderStatusTypeFilled)

	return
}
