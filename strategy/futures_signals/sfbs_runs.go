package futures_signals

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	types "github.com/fr0ster/go-trading-utils/types"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	deltaUp    = 0.0005
	deltaDown  = 0.0005
	degree     = 3
	limit      = 1000
	interval   = "1m"
	reloadTime = 500 * time.Millisecond
)

func RunFuturesHolding(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	stopEvent <- os.Interrupt
	return fmt.Errorf("it should be implemented for futures")
}

func RunScalpingHolding(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal) (err error) {
	pair.SetStrategy(pairs_types.GridStrategyType)
	return RunFuturesGridTrading(config, client, pair, stopEvent)
}

func RunFuturesTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.ScalpingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}

	if config.GetConfigurations().GetReloadConfig() {
		go func() {
			for {
				<-time.After(reloadTime)
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
			}
		}()
	}

	stopEvent <- os.Interrupt
	return fmt.Errorf("it hadn't been implemented yet")
}

// Створення ордера для розміщення в грід
func createOrderInGrid(
	pairProcessor *PairProcessor,
	side futures.SideType,
	quantity,
	price float64) (order *futures.CreateOrderResponse, err error) {
	order, err = pairProcessor.CreateOrder(
		futures.OrderTypeLimit,     // orderType
		side,                       // sideType
		futures.TimeInForceTypeGTC, // timeInForce
		quantity,                   // quantity
		false,                      // closePosition
		price,                      // price
		0,                          // stopPrice
		0)                          // callbackRate
	if err == nil && order != nil {
		logrus.Errorf("Futures %s: Side %v, TimeInForce %v, Quantity %v, Price %v",
			order.Symbol, order.Side, order.TimeInForce, order.OrigQuantity, order.Price)
	} else {
		logrus.Errorf("Futures: Side %s, TimeInForce %s, Quantity %v, Price %v, error %s",
			side, futures.TimeInForceTypeGTC, quantity, price, err)
	}
	return
}

// Округлення ціни до TickSize знаків після коми
func roundPrice(val float64, symbol *futures.Symbol) float64 {
	exp := int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
	return utils.RoundToDecimalPlace(val, exp)
}

func IsOrdersOpened(grid *grid_types.Grid, pairProcessor *PairProcessor, pair *pairs_types.Pairs) (err error) {
	grid.Ascend(func(item btree.Item) bool {
		var orderOut *futures.Order
		record := item.(*grid_types.Record)
		if record.GetOrderId() != 0 {
			orderOut, err = pairProcessor.GetOrder(record.GetOrderId())
			if err != nil {
				return false
			}
			if orderOut == nil || orderOut.Status != futures.OrderStatusTypeNew {
				err = fmt.Errorf("futures %s: Order %v not found or status %v", pair.GetPair(), record.GetOrderId(), orderOut.Status)
			}
		}
		return true
	})
	return err
}

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams,
	symbol *futures.Symbol,
	side futures.SideType,
	grid *grid_types.Grid,
	order *grid_types.Record,
	quantity float64) (err error) {
	var (
		takerPrice *grid_types.Record
		takerOrder *futures.CreateOrderResponse
	)
	if side == futures.SideTypeSell {
		// Якшо вище немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetUpPrice() == 0 {
			// Створюємо ордер на продаж
			price := roundPrice(order.GetPrice()*(1+pair.GetSellDelta()), symbol)
			distance := pairStreams.GetLiquidationDistance(price)
			if (pair.GetUpBound() == 0 || price <= pair.GetUpBound()) && distance > config.GetConfigurations().GetPercentsToLiquidation() {
				upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, price)
				if err != nil {
					return err
				}
				logrus.Debugf("Futures %s: Set Sell order %v on price %v status %v quantity %v",
					pair.GetPair(), upOrder.OrderID, price, upOrder.Status, quantity)
				// Записуємо ордер в грід
				upPrice := grid_types.NewRecord(upOrder.OrderID, price, 0, order.GetPrice(), types.OrderSide(futures.SideTypeSell))
				grid.Set(upPrice)
				order.SetUpPrice(price) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status != futures.OrderStatusTypeNew {
					takerPrice = upPrice
					takerOrder = upOrder
				}
			} else {
				logrus.Debugf("Futures %s: UpBound %v, price %v, distance %v more than %v",
					pair.GetPair(), pair.GetUpBound(), price, distance, config.GetConfigurations().GetPercentsToLiquidation())
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель нижче
		downPrice, ok := grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
		if ok && downPrice.GetOrderId() == 0 {
			// Створюємо ордер на купівлю
			downOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, order.GetDownPrice())
			if err != nil {
				return err
			}
			downPrice.SetOrderId(downOrder.OrderID)   // Записуємо номер ордера в грід
			downPrice.SetOrderSide(types.SideTypeBuy) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: Set Buy order %v on price %v status %v quantity %v",
				pair.GetPair(), downOrder.OrderID, order.GetDownPrice(), downOrder.Status, quantity)
			if downOrder.Status != futures.OrderStatusTypeNew {
				takerPrice = downPrice
				takerOrder = downOrder
			}
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair, pairStreams,
				symbol,
				takerOrder.Side,
				grid,
				takerPrice,
				quantity)
			if err != nil {
				return err
			}
		}
	} else if side == futures.SideTypeBuy {
		// Якшо нижче немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetDownPrice() == 0 {
			// Створюємо ордер на купівлю
			price := roundPrice(order.GetPrice()*(1-pair.GetBuyDelta()), symbol)
			distance := pairStreams.GetLiquidationDistance(price)
			if (pair.GetLowBound() == 0 || price >= pair.GetLowBound()) && distance > config.GetConfigurations().GetPercentsToLiquidation() {
				downOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, price)
				if err != nil {
					return err
				}
				logrus.Debugf("Futures %s: Set Buy order %v on price %v status %v quantity %v",
					pair.GetPair(), downOrder.OrderID, price, downOrder.Status, quantity)
				// Записуємо ордер в грід
				downPrice := grid_types.NewRecord(downOrder.OrderID, price, order.GetPrice(), 0, types.OrderSide(futures.SideTypeBuy))
				grid.Set(downPrice)
				order.SetDownPrice(price) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status != futures.OrderStatusTypeNew {
					takerPrice = downPrice
					takerOrder = downOrder
				}
			} else {
				logrus.Debugf("Futures %s: LowBound %v, price %v, distance %v less than %v",
					pair.GetPair(), pair.GetLowBound(), price, distance, config.GetConfigurations().GetPercentsToLiquidation())
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель вище
		upPrice, ok := grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		if ok && upPrice.GetOrderId() == 0 {
			// Створюємо ордер на продаж
			upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, order.GetUpPrice())
			if err != nil {
				logrus.Errorf("Futures %s: Set Sell order %v on price %v status %v quantity %v",
					pair.GetPair(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status, quantity)
				return err
			}
			if upOrder.Status != futures.OrderStatusTypeNew {
				takerPrice = upPrice
				takerOrder = upOrder
			}
			upPrice.SetOrderId(upOrder.OrderID)      // Записуємо номер ордера в грід
			upPrice.SetOrderSide(types.SideTypeSell) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: Set Sell order %v on price %v status %v quantity %v",
				pair.GetPair(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status, quantity)
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair, pairStreams,
				symbol,
				takerOrder.Side,
				grid,
				takerPrice,
				quantity)
			if err != nil {
				return err
			}
		}
	}
	return
}

func updateConfig(
	config *config_types.ConfigFile,
	pairStreams *PairStreams,
	pair *pairs_types.Pairs,
	price float64) (quantity float64, err error) {
	// Обновляємо конфігурацію
	if config.GetConfigurations().GetReloadConfig() {
		config.Load()
		pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
		balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
		if err != nil {
			return 0, err
		}
		pair.SetCurrentBalance(balance)
		config.Save()
		symbol, err := func() (res *futures.Symbol, err error) {
			val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair.GetPair()})
			if val == nil {
				return nil, fmt.Errorf("spot %s: Symbol not found", pair.GetPair())
			}
			return val.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
		}()
		if err != nil {
			return 0, err
		}
		logrus.Debugf("Futures %s: symbol %v", pair.GetPair(), symbol)
		quantity = pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() * float64(pair.GetLeverage()) / price
		minNotional := utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
		if quantity*price < minNotional {
			logrus.Debugf("Futures %s: Quantity %v * price %v < minNotional %v", pair.GetPair(), quantity, price, minNotional)
			quantity = minNotional / price
		}
	}
	return
}

func balancingMargin(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pairStreams *PairStreams,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	event *futures.WsUserDataEvent) (err error) {
	distance := pairStreams.GetLiquidationDistance(utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice))
	if utils.ConvStrToFloat64(risk.LiquidationPrice) != 0 && distance <= config.GetConfigurations().GetPercentsToLiquidation() {
		if utils.ConvStrToFloat64(risk.PositionAmt) != 0 && utils.ConvStrToFloat64(risk.IsolatedMargin) != 0 && utils.ConvStrToFloat64(risk.IsolatedMargin) < pair.GetCurrentPositionBalance() {
			err = pairProcessor.SetPositionMargin(pair.GetCurrentPositionBalance(), 1)
			if err != nil {
				return err
			}
			logrus.Debugf("Futures %s: Margin was %v, add Margin %v, new Margin %v",
				pair.GetPair(),
				utils.ConvStrToFloat64(risk.IsolatedMargin),
				pair.GetCurrentPositionBalance(),
				pairProcessor.GetPositionMargin())
		}
	} else if utils.ConvStrToFloat64(risk.LiquidationPrice) == 0 {
		if utils.ConvStrToFloat64(risk.PositionAmt) != 0 && utils.ConvStrToFloat64(risk.IsolatedMargin) != 0 && utils.ConvStrToFloat64(risk.IsolatedMargin) > pair.GetCurrentPositionBalance() {
			err = pairProcessor.SetPositionMargin(utils.ConvStrToFloat64(risk.IsolatedMargin)-pair.GetCurrentPositionBalance(), 2)
			if err != nil {
				return err
			}
			logrus.Debugf("Futures %s: Margin was %v, remove Margin %v, new Margin %v",
				pair.GetPair(),
				utils.ConvStrToFloat64(risk.IsolatedMargin),
				utils.ConvStrToFloat64(risk.IsolatedMargin)-pair.GetCurrentPositionBalance(),
				pairProcessor.GetPositionMargin())
		}
	}
	return
}

func observePriceLiquidation(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams,
	event *futures.WsUserDataEvent) (err error) {
	if config.GetConfigurations().GetObservePriceLiquidation() {
		price := utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
		risk, err := pairStreams.GetPositionRisk()
		if err != nil {
			return err
		}
		delta_percent := pairStreams.GetLiquidationDistance(price)
		if delta_percent <= config.GetConfigurations().GetPercentsToLiquidation() {
			logrus.Debugf("Futures %s: Liquidation price %v, delta %v!!!!!", pair.GetPair(), risk.LiquidationPrice, delta_percent)
			free, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
			if err != nil {
				return err
			}
			if free >= pair.GetCurrentPositionBalance() {
				logrus.Debugf("Futures %s: Free asset %v >= current balance %v", pair.GetPair(), free, pair.GetCurrentPositionBalance())
				// Устанавлюемо Margin
				err = pairProcessor.SetPositionMargin(pair.GetCurrentPositionBalance(), 1)
				if err != nil {
					return err
				}
			} else {
				logrus.Debugf("Futures %s: Free asset %v < current balance %v", pair.GetPair(), free, pair.GetCurrentPositionBalance())
				positionAmtDec := utils.ConvStrToFloat64(risk.PositionAmt) * delta_percent
				if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
					logrus.Debugf("Futures %s: Liquidation price %v, delta %v, position %v, new position %v",
						pair.GetPair(), risk.LiquidationPrice, delta_percent, risk.PositionAmt, positionAmtDec)
					_, err = pairProcessor.CreateOrder(
						futures.OrderTypeMarket,    // orderType
						futures.SideTypeSell,       // sideType
						futures.TimeInForceTypeGTC, // timeInForce
						positionAmtDec,             // quantity
						false,                      // closePosition
						0,                          // price
						0,                          // stopPrice
						0)                          // callbackRate
				} else if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
					logrus.Debugf("Futures %s: Liquidation price %v, delta %v, position %v, new position %v",
						pair.GetPair(), risk.LiquidationPrice, delta_percent, risk.PositionAmt, positionAmtDec)
					_, err = pairProcessor.CreateOrder(
						futures.OrderTypeMarket,    // orderType
						futures.SideTypeBuy,        // sideType
						futures.TimeInForceTypeGTC, // timeInForce
						positionAmtDec,             // quantity
						false,                      // closePosition
						0,                          // price
						0,                          // stopPrice
						0)                          // callbackRate
				}
				return err
			}
		}
	}
	return
}

func RunFuturesGridTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal) (err error) {
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		stopEvent <- os.Interrupt
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.GridStrategyType {
		stopEvent <- os.Interrupt
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		stopEvent <- os.Interrupt
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	// Створюємо стрім подій
	pairStreams, err := NewPairStreams(client, pair, false)
	if err != nil {
		stopEvent <- os.Interrupt
		return
	}
	// Створюємо обробник пари
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), false)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}

	balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	pair.SetCurrentBalance(balance)
	config.Save()
	if pair.GetInitialBalance() == 0 {
		pair.SetInitialBalance(balance)
		config.Save()
	}
	if pair.GetMarginType() == "" {
		logrus.Debugf("Futures %s set MarginType %v from account into config", pair.GetPair(), pairProcessor.GetMarginType())
		pair.SetMarginType(pairProcessor.GetMarginType())
		config.Save()
	} else {
		if pair.GetMarginType() != pairProcessor.GetMarginType() {
			logrus.Debugf("Futures %s set MarginType %v from config into account", pair.GetPair(), pair.GetMarginType())
			pairProcessor.SetMarginType(pair.GetMarginType())
		}
	}
	if pair.GetLeverage() == 0 {
		logrus.Debugf("Futures %s set Leverage %v from account into config", pair.GetPair(), pairProcessor.GetLeverage())
		pair.SetLeverage(pairProcessor.GetLeverage())
		config.Save()
	} else {
		if pair.GetLeverage() != pairProcessor.GetLeverage() {
			logrus.Debugf("Futures %s set Leverage %v from config into account", pair.GetPair(), pair.GetLeverage())
			pairProcessor.SetLeverage(pair.GetLeverage())
		}
	}
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pair.GetPair())
	grid := grid_types.New()
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		stopEvent <- os.Interrupt
		return fmt.Errorf("futures %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
	}
	symbol, err := func() (res *futures.Symbol, err error) {
		val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair.GetPair()})
		if val == nil {
			return nil, fmt.Errorf("futures %s: Symbol not found", pair.GetPair())
		}
		return val.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
	}()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Отримання середньої ціни
	price := roundPrice(pair.GetMiddlePrice(), symbol)
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	if entryPrice := utils.ConvStrToFloat64(risk.EntryPrice); entryPrice != 0 {
		price = roundPrice(entryPrice, symbol)
	}
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = roundPrice(price, symbol)
	}
	quantity := pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() * float64(pair.GetLeverage()) / price
	minNotional := utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
	logrus.Debugf("Futures %s: minNotional %v", pair.GetPair(), minNotional)
	if quantity*price < minNotional {
		logrus.Debugf("Futures %s: Quantity %v * price %v < minNotional %v", pair.GetPair(), quantity, price, minNotional)
		quantity = minNotional / price
	}
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, roundPrice(price*(1+pair.GetSellDelta()), symbol), roundPrice(price*(1-pair.GetBuyDelta()), symbol), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pair.GetPair(), price)

	err = pairProcessor.CancelAllOrders()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Створюємо ордери на продаж
	sellOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, roundPrice(price*(1+pair.GetSellDelta()), symbol))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, roundPrice(price*(1+pair.GetSellDelta()), symbol), 0, price, types.SideTypeSell))
	// logrus.Debugf("Futures %s: Set Sell order on price %v", pair.GetPair(), roundPrice(price*(1+pair.GetSellDelta()), symbol))
	// Створюємо ордер на купівлю
	buyOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, roundPrice(price*(1-pair.GetBuyDelta()), symbol))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, roundPrice(price*(1-pair.GetSellDelta()), symbol), price, 0, types.SideTypeBuy))
	// logrus.Debugf("Futures %s: Set Buy order on price %v", pair.GetPair(), roundPrice(price*(1-pair.GetBuyDelta()), symbol))
	// Стартуємо обробку ордерів
	// grid.Debug("Futures Grid", "", pair.GetPair())
	// logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	for {
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			grid.Lock()
			price := utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
			// Обновляємо конфігурацію
			quantity, err := updateConfig(config, pairStreams, pair, price)
			if err != nil {
				return err
			}
			// Зміна маржі при потребі
			err = balancingMargin(config, pairProcessor, pairStreams, pair, risk, event)
			if err != nil {
				return err
			}
			// Спостереження за ліквідацією при потребі
			err = observePriceLiquidation(config, pairProcessor, pair, pairStreams, event)
			if err != nil {
				return err
			}
			logrus.Debugf("Futures %s: Order %v on price %v side %v status %s",
				pair.GetPair(),
				event.OrderTradeUpdate.ID,
				event.OrderTradeUpdate.OriginalPrice,
				event.OrderTradeUpdate.Side,
				event.OrderTradeUpdate.Status)
			// Знаходимо у гріді на якому був виконаний ордер
			order, ok := grid.Get(&grid_types.Record{Price: utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)}).(*grid_types.Record)
			if !ok {
				return fmt.Errorf("uncorrected order ID: %v", event.OrderTradeUpdate.ID)
			}
			orderId := order.GetOrderId()
			err = processOrder(config, pairProcessor, pair, pairStreams, symbol, event.OrderTradeUpdate.Side, grid, order, quantity)
			if err != nil {
				pairProcessor.CancelAllOrders()
				return err
			}
			grid.Unlock()
			grid.Debug("Futures Grid After processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
			pairProcessor.Debug("Futures Pair After processOrder", strconv.FormatInt(orderId, 10))
		case <-time.After(60 * time.Second):
			grid.Debug("Futures Grid", "", pair.GetPair())
		}
	}
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	debug bool) (err error) {
	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		return RunFuturesHolding(config, client, degree, limit, pair, stopEvent, time.Second, debug)

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		return RunScalpingHolding(config, client, pair, stopEvent)

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		return RunFuturesTrading(config, client, degree, limit, pair, stopEvent, time.Second, debug)

		// Відпрацьовуємо Grid стратегію
	} else if pair.GetStrategy() == pairs_types.GridStrategyType {
		return RunFuturesGridTrading(config, client, pair, stopEvent)

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		return fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
	}
}
