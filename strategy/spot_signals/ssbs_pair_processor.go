package spot_signals

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
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
		config       *config_types.ConfigFile
		client       *binance.Client
		pair         pairs_interfaces.Pairs
		exchangeInfo *exchange_types.ExchangeInfo
		account      *spot_account.Account

		buyEvent            chan *pair_price_types.PairPrice
		buyProcessRun       bool
		stopBuy             chan bool
		startBuyOrderEvent  chan *binance.CreateOrderResponse
		sellEvent           chan *pair_price_types.PairPrice
		sellProcessRun      bool
		startSellOrderEvent chan *binance.CreateOrderResponse
		stopSell            chan bool

		startProcessBuyTakeProfitOrderEvent  chan *binance.CreateOrderResponse
		buyTakeProfitProcessRun              bool
		stopBuyTakeProfitProcess             chan bool
		startProcessSellTakeProfitOrderEvent chan *binance.CreateOrderResponse
		sellTakeProfitProcessRun             bool
		stopSellTakeProfitProcess            chan bool

		orderExecuted                  chan bool
		stopOrderExecutionGuardProcess chan bool
		orderExecutionGuardProcessRun  bool

		userDataEvent    chan *binance.WsUserDataEvent
		orderStatusEvent chan *binance.WsUserDataEvent

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

		stop      chan os.Signal
		limitsOut chan bool

		pairInfo *symbol_types.SpotSymbol
		degree   int
		debug    bool
	}
)

//  1. LIMIT_MAKER are LIMIT orders that will be rejected if they would immediately match and trade as a taker.
//  2. STOP_LOSS and TAKE_PROFIT will execute a MARKET order when the stopPrice is reached.
//     Any LIMIT or LIMIT_MAKER type order can be made an iceberg order by sending an icebergQty.
//     Any order with an icebergQty MUST have timeInForce set to GTC.
//  3. MARKET orders using the quantity field specifies the amount of the base asset the user wants to buy or sell at the market price.
//     For example, sending a MARKET order on BTCUSDT will specify how much BTC the user is buying or selling.
//  4. MARKET orders using quoteOrderQty specifies the amount the user wants to spend (when buying) or receive (when selling) the quote asset;
//     the correct quantity will be determined based on the market liquidity and quoteOrderQty.
//     Using BTCUSDT as an example:
//     On the BUY side, the order will buy as many BTC as quoteOrderQty USDT can.
//     On the SELL side, the order will sell as much BTC needed to receive quoteOrderQty USDT.
//  5. MARKET orders using quoteOrderQty will not break LOT_SIZE filter rules; the order will execute a quantity that will have the notional value as close as possible to quoteOrderQty.
//     same newClientOrderId can be accepted only when the previous one is filled, otherwise the order will be rejected.
//  6. For STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT and TAKE_PROFIT orders, trailingDelta can be combined with stopPrice.
//
//  7. Trigger order price rules against market price for both MARKET and LIMIT versions:
//     Price above market price: STOP_LOSS BUY, TAKE_PROFIT SELL
//     Price below market price: STOP_LOSS SELL, TAKE_PROFIT BUY
func (pp *PairProcessor) CreateOrder(
	orderType binance.OrderType, // MARKET, LIMIT, LIMIT_MAKER, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	sideType binance.SideType, // BUY, SELL
	timeInForce binance.TimeInForceType, // GTC, IOC, FOK
	quantity float64, // BTC for example if we buy or sell BTC
	quantityQty float64, // USDT for example if we buy or sell BTC
	// price for 1 BTC
	// it's price of order execution for LIMIT, LIMIT_MAKER
	// after execution of STOP_LOSS, TAKE_PROFIT, wil be created MARKET order
	// after execution of STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT wil be created LIMIT order with price of order execution from PRICE parameter
	price float64,
	stopPrice float64, // price for stop loss or take profit it's price of order execution for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	trailingDelta int) ( // trailingDelta for stop loss or take profit
	order *binance.CreateOrderResponse, err error) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
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
			Symbol(string(binance.SymbolType(pp.pair.GetPair()))).
			Type(orderType).
			Side(sideType)
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == binance.OrderTypeMarket {
		// MARKET	quantity or quoteOrderQty
		if quantity != 0 {
			service = service.
				Quantity(utils.ConvFloat64ToStr(quantity, quantityRound))
		} else if quantityQty != 0 {
			service = service.
				QuoteOrderQty(utils.ConvFloat64ToStr(quantityQty, quantityRound))
		} else {
			err = fmt.Errorf("quantity or quoteOrderQty must be set")
			return
		}
	} else if orderType == binance.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
	} else if orderType == binance.OrderTypeLimitMaker {
		// LIMIT_MAKER	quantity, price
		service = service.
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
	} else if orderType == binance.OrderTypeStopLoss || orderType == binance.OrderTypeTakeProfit {
		// STOP_LOSS	quantity, stopPrice or trailingDelta
		service = service.
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound))
		if stopPrice != 0 {
			service = service.StopPrice(utils.ConvFloat64ToStr(price, priceRound))
		} else if trailingDelta != 0 {
			service = service.TrailingDelta(strconv.Itoa(trailingDelta))
		} else {
			err = fmt.Errorf("stopPrice or trailingDelta must be set")
			return
		}
	} else if orderType == binance.OrderTypeStopLossLimit || orderType == binance.OrderTypeTakeProfitLimit {
		// STOP_LOSS_LIMIT	timeInForce, quantity, price, stopPrice or trailingDelta
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
		if stopPrice != 0 {
			service = service.StopPrice(utils.ConvFloat64ToStr(price, priceRound))
		} else if trailingDelta != 0 {
			service = service.TrailingDelta(strconv.Itoa(trailingDelta))
		} else {
			err = fmt.Errorf("stopPrice or trailingDelta must be set")
			return
		}
	}
	return service.Do(context.Background())
}

func (pp *PairProcessor) ProcessBuyOrder() (nextTriggerEvent chan *binance.CreateOrderResponse, err error) {
	if !pp.buyProcessRun {
		if pp.startBuyOrderEvent == nil {
			pp.startBuyOrderEvent = make(chan *binance.CreateOrderResponse)
		}
		nextTriggerEvent = pp.startBuyOrderEvent
		go func() {
			var order *binance.CreateOrderResponse
			for {
				select {
				case <-pp.stopBuy:
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
					baseBalance, err := GetBaseBalance(pp.account, pp.pair)
					if err != nil {
						logrus.Errorf("Can't get %s asset: %v", pp.pair.GetBaseSymbol(), err)
						pp.stop <- os.Interrupt
						return
					}
					if baseBalance < params.Quantity*params.Price {
						logrus.Warnf("We don't buy, we need %v of %s for buy %v of %s",
							baseBalance, pp.pair.GetTargetSymbol(), params.Quantity*params.Price, pp.pair.GetTargetSymbol())
						continue
					}
					if !pp.debug {
						order, err = pp.CreateOrder(
							binance.OrderTypeMarket,
							binance.SideTypeBuy,
							binance.TimeInForceTypeGTC,
							params.Quantity,
							0,
							params.Price,
							0,
							0)
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
							nextTriggerEvent <- order
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
		pp.buyProcessRun = true
	} else {
		nextTriggerEvent = pp.startBuyOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellOrder() (nextTriggerEvent chan *binance.CreateOrderResponse, err error) {
	if !pp.sellProcessRun {
		if pp.startSellOrderEvent == nil {
			pp.startSellOrderEvent = make(chan *binance.CreateOrderResponse)
		}
		nextTriggerEvent = pp.startSellOrderEvent
		go func() {
			var order *binance.CreateOrderResponse
			for {
				select {
				case <-pp.stopSell:
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
						order, err = pp.CreateOrder(
							binance.OrderTypeMarket,
							binance.SideTypeBuy,
							binance.TimeInForceTypeGTC,
							params.Quantity,
							0,
							params.Price,
							0,
							0)
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
							nextTriggerEvent <- order
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
		pp.sellProcessRun = true
	} else {
		nextTriggerEvent = pp.startSellOrderEvent
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

func (pp *PairProcessor) ProcessBuyTakeProfitOrder(trailingDelta int) (nextTriggerEvent chan *binance.CreateOrderResponse) {
	if !pp.buyTakeProfitProcessRun {
		if pp.startProcessBuyTakeProfitOrderEvent == nil {
			pp.startProcessBuyTakeProfitOrderEvent = make(chan *binance.CreateOrderResponse)
		}
		nextTriggerEvent = pp.startProcessBuyTakeProfitOrderEvent
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
							binance.OrderTypeTakeProfit,
							binance.SideTypeBuy,
							binance.TimeInForceTypeGTC,
							params.Quantity,
							0, // We don't use quoteOrderQty
							params.Price,
							0, // We don't use stopPrice
							trailingDelta)
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
							nextTriggerEvent <- order
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
		pp.buyTakeProfitProcessRun = true
	} else {
		nextTriggerEvent = pp.startProcessBuyTakeProfitOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellTakeProfitOrder(trailingDelta int) (nextTriggerEvent chan *binance.CreateOrderResponse) {
	if !pp.sellTakeProfitProcessRun {
		if pp.startProcessSellTakeProfitOrderEvent == nil {
			pp.startProcessSellTakeProfitOrderEvent = make(chan *binance.CreateOrderResponse)
		}
		nextTriggerEvent = pp.startProcessSellTakeProfitOrderEvent
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
							binance.OrderTypeTakeProfit,
							binance.SideTypeBuy,
							binance.TimeInForceTypeGTC,
							params.Quantity,
							0, // We don't use quoteOrderQty
							params.Price,
							0, // We don't use stopPrice
							trailingDelta)
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
							nextTriggerEvent <- order
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
		pp.sellTakeProfitProcessRun = true
	} else {
		nextTriggerEvent = pp.startProcessSellTakeProfitOrderEvent
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

func (pp *PairProcessor) ProcessAfterBuyOrder(triggerEvent chan *binance.CreateOrderResponse) {
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

func (pp *PairProcessor) ProcessAfterSellOrder(triggerEvent chan *binance.CreateOrderResponse) {
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
	if !pp.orderExecutionGuardProcessRun {
		if pp.orderExecuted == nil {
			pp.orderExecuted = make(chan bool, 1)
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

func NewPairProcessor(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair pairs_interfaces.Pairs,
	// orderType binance.OrderType,
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice,
	debug bool) (pp *PairProcessor, err error) {
	pp = &PairProcessor{
		client:    client,
		pair:      pair,
		account:   nil,
		stop:      make(chan os.Signal, 1),
		limitsOut: make(chan bool, 1),
		pairInfo:  nil,

		buyEvent:                 buyEvent,
		buyProcessRun:            false,
		sellEvent:                sellEvent,
		sellProcessRun:           false,
		buyTakeProfitProcessRun:  false,
		sellTakeProfitProcessRun: false,

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		stopBuy:                        make(chan bool, 1),
		stopSell:                       make(chan bool, 1),
		stopOrderExecutionGuardProcess: make(chan bool, 1),

		orderExecuted:    nil,
		orderStatusEvent: nil,
		degree:           3,
		debug:            debug,
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
