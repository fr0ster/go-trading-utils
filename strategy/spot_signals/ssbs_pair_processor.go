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
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairProcessor struct {
		config       *config_types.ConfigFile
		client       *binance.Client
		pair         *pairs_types.Pairs
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

		pairInfo     *symbol_types.SpotSymbol
		orderTypes   map[string]bool
		degree       int
		debug        bool
		sleepingTime time.Duration
		timeOut      time.Duration
	}
)

func (pp *PairProcessor) GetBaseBalance() (
	baseBalance float64, // Кількість базової валюти
	err error) {
	baseBalance, err = func() (
		baseBalance float64,
		err error) {
		baseBalance, err = pp.account.GetFreeAsset(pp.pair.GetBaseSymbol())
		return
	}()

	if err != nil {
		return 0, err
	}
	return
}

func (pp *PairProcessor) GetTargetBalance() (
	targetBalance float64, // Кількість торгової валюти
	err error) {
	targetBalance, err = func() (
		targetBalance float64,
		err error) {
		targetBalance, err = pp.account.GetFreeAsset(pp.pair.GetTargetSymbol())
		return
	}()

	if err != nil {
		return 0, err
	}
	return
}

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
	// price for stop loss or take profit it's price of order execution for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	stopPrice float64,
	// trailingDelta for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	// https://github.com/binance/binance-spot-api-docs/blob/master/faqs/trailing-stop-faq.md
	trailingDelta int) (
	order *binance.CreateOrderResponse, err error) {
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	if _, ok := pp.orderTypes[string(orderType)]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.pair.GetPair())
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderRespType(binance.NewOrderRespTypeRESULT).
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
		// STOP_LOSS/TAKE_PROFIT quantity, stopPrice or trailingDelta
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
		// STOP_LOSS_LIMIT/TAKE_PROFIT_LIMIT timeInForce, quantity, price, stopPrice or trailingDelta
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

func (pp *PairProcessor) ProcessBuyOrder(triggerEvent chan *pair_price_types.PairPrice) (nextTriggerEvent chan *binance.CreateOrderResponse, err error) {
	if !pp.buyProcessRun {
		if pp.buyEvent == nil {
			pp.buyEvent = triggerEvent
		}
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
				time.Sleep(pp.sleepingTime)
			}
		}()
		pp.buyProcessRun = true
	} else {
		nextTriggerEvent = pp.startBuyOrderEvent
	}
	return
}

func (pp *PairProcessor) ProcessSellOrder(triggerEvent chan *pair_price_types.PairPrice) (nextTriggerEvent chan *binance.CreateOrderResponse, err error) {
	if !pp.sellProcessRun {
		if pp.sellEvent == nil {
			pp.sellEvent = triggerEvent
		}
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
				time.Sleep(pp.sleepingTime)
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
				time.Sleep(pp.sleepingTime)
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
				time.Sleep(pp.sleepingTime)
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

func (pp *PairProcessor) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairProcessor) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func (pp *PairProcessor) CheckOrderType(orderType binance.OrderType) bool {
	_, ok := pp.orderTypes[string(orderType)]
	return ok
}

func (pp *PairProcessor) GetOpenOrders() (orders []*binance.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*binance.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *binance.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.pair.GetPair()).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *binance.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.pair.GetPair()).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (orders *binance.CancelOpenOrdersResponse, err error) {
	return pp.client.NewCancelOpenOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetUserDataEvent() chan *binance.WsUserDataEvent {
	return pp.userDataEvent
}

func (pp *PairProcessor) GetOrderStatusEvent() chan *binance.WsUserDataEvent {
	return pp.orderStatusEvent
}

func (pp *PairProcessor) GetPair() *pairs_types.Pairs {
	return pp.pair
}

func (pp *PairProcessor) Debug(fl string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		orders, _ := pp.GetOpenOrders()
		logrus.Debugf("%s: Open orders for %s", fl, pp.pair.GetPair())
		for _, order := range orders {
			logrus.Debugf(" Order %v on price %v OrderSide %v", order.OrderID, order.Price, order.Side)
		}
	}
}

func NewPairProcessor(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *pairs_types.Pairs,
	exchangeInfo *exchange_types.ExchangeInfo,
	account *spot_account.Account,
	userDataEvent chan *binance.WsUserDataEvent,
	debug bool) (pp *PairProcessor, err error) {
	pp = &PairProcessor{
		config:       config,
		client:       client,
		pair:         pair,
		exchangeInfo: exchangeInfo,
		account:      account,

		buyEvent:            nil,
		buyProcessRun:       false,
		stopBuy:             nil,
		startBuyOrderEvent:  nil,
		sellEvent:           nil,
		sellProcessRun:      false,
		startSellOrderEvent: nil,
		stopSell:            nil,

		buyTakeProfitProcessRun:  false,
		sellTakeProfitProcessRun: false,

		orderExecuted:                  nil,
		stopOrderExecutionGuardProcess: nil,
		orderExecutionGuardProcessRun:  false,

		orderStatusEvent: nil,

		userDataEvent: userDataEvent,

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		stop:      make(chan os.Signal, 1),
		limitsOut: make(chan bool, 1),

		pairInfo:     nil,
		orderTypes:   map[string]bool{},
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
		LimitRead(degree, []string{pp.pair.GetPair()}, client)

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.SpotSymbol{Symbol: pair.GetPair()}).(*symbol_types.SpotSymbol)

	// Ініціалізуємо типи ордерів які можна використовувати для пари
	pp.orderTypes = make(map[string]bool, 0)
	for _, orderType := range pp.pairInfo.OrderTypes {
		pp.orderTypes[orderType] = true
	}

	// Ініціалізуємо стріми для оновлення лімітів на ордери та запити
	pp.LimitUpdaterStream()

	// Визначаємо статуси ордерів які нас цікавлять ...
	// ... запускаємо стрім для відслідковування зміни статусу ордерів які нас цікавлять
	if config.GetConfigurations().GetMaintainPartiallyFilledOrders() {
		pp.orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(pp.userDataEvent, binance.OrderStatusTypeFilled, binance.OrderStatusTypePartiallyFilled)
	} else {
		pp.orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(pp.userDataEvent, binance.OrderStatusTypeFilled)
	}

	return
}
