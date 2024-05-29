package spot_signals

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"

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
	reloadTime = 500 * time.Millisecond
)

func RunSpotHolding(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage || pair.GetStage() == pairs_types.OutputOfPositionStage {
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

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	buyEvent, _ := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)

	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-time.After(updateTime):
				triggerEvent <- true
			}
		}
	}()

	collectionOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)

	pairStream, err := NewPairStreams(client, pair, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	_, err = pairProcessor.ProcessBuyOrder(buyEvent)
	if err != nil {
		return err
	}

	<-collectionOutEvent
	pairProcessor.StopBuySignal()
	pair.SetStage(pairs_types.PositionClosedStage)
	config.Save()
	stopEvent <- os.Interrupt
	return nil
}

func RunSpotScalping(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
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

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}

	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-sellEvent:
				triggerEvent <- true
			}
		}
	}()

	pairStream, err := NewPairStreams(client, pair, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessBuyOrder(buyEvent)
		if err != nil {
			return err
		}
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
		<-collectionOutEvent
		_, err = pairProcessor.ProcessSellOrder(sellEvent)
		if err != nil {
			return err
		}
		pair.SetStage(pairs_types.WorkInPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessSellOrder(sellEvent) // Все одно другий раз не запустится, бо вже працює горутина
		if err != nil {
			return err
		}
		workingOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)
		_, err = pairProcessor.ProcessSellOrder(sellEvent)
		if err != nil {
			return err
		}

		<-workingOutEvent
		pairProcessor.StopBuySignal()
		pair.SetStage(pairs_types.OutputOfPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		pairProcessor.StopBuySignal() // Зупиняємо купівлю, продаємо поки є шо продавати
		if err != nil {
			return err
		}
		positionClosed := pairObserver.ClosePositionSignal(triggerEvent) // Чекаємо на закриття позиції
		<-positionClosed
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
	}
	return nil
}

func RunSpotTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.TradingStrategyType {
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

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}

	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-sellEvent:
				triggerEvent <- true
			}
		}
	}()

	pairStream, err := NewPairStreams(client, pair, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeTakeProfit) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeTakeProfitLimit)
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessBuyOrder(buyEvent)
		if err != nil {
			return err
		}
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
		<-collectionOutEvent
		pair.SetStage(pairs_types.OutputOfPositionStage) // В trading стратегії не спекулюємо, накопили позицію і закриваемо продажем лімітним ордером
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		pairProcessor.StopBuySignal() // Зупиняємо купівлю, продаємо поки є шо продавати
		// TODO: Закриття позиції лімітним trailing ордером
		quantity, err := GetTargetBalance(account, pair)
		if err != nil {
			return err
		}
		order, err := pairProcessor.CreateOrder(
			binance.OrderTypeTakeProfitLimit,
			binance.SideTypeSell,
			binance.TimeInForceTypeGTC,
			// STOP_LOSS_LIMIT/TAKE_PROFIT_LIMIT timeInForce, quantity, price, stopPrice or trailingDelta
			quantity,
			0,   // quantityQty
			0,   // price
			0,   // stopPrice
			100) // trailingDelta
		if err != nil {
			return err
		}
		positionClosed := pairProcessor.OrderExecutionGuard(order) // Чекаємо на закриття позиції
		<-positionClosed
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
	}
	return nil
}

// Створення ордера для розміщення в грід
func createOrderInGrid(
	pairProcessor *PairProcessor,
	side binance.SideType,
	quantity,
	price float64) (order *binance.CreateOrderResponse, err error) {
	order, err = pairProcessor.CreateOrder(
		binance.OrderTypeLimit,     // orderType
		side,                       // sideType
		binance.TimeInForceTypeGTC, // timeInForce
		quantity,                   // quantity
		0,                          // quantityQty
		price,                      // price
		0,                          // stopPrice
		0)                          // trailingDelta
	return
}

// Округлення ціни до TickSize знаків після коми
func roundPrice(val float64, symbol *binance.Symbol) float64 {
	exp := int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
	return utils.RoundToDecimalPlace(val, exp)
}

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams,
	symbol *binance.Symbol,
	side binance.SideType,
	grid *grid_types.Grid,
	order *grid_types.Record,
	quantity float64) (err error) {
	var (
		takerPrice *grid_types.Record
		takerOrder *binance.CreateOrderResponse
	)
	grid.Debug("Spots Grid Before processOrder", strconv.FormatInt(order.OrderId, 10), pair.GetPair())
	pairProcessor.Debug("Spots Pair Before processOrder")
	if side == binance.SideTypeSell {
		// Якшо вище немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetUpPrice() == 0 {
			// Створюємо ордер на продаж
			price := roundPrice(order.GetPrice()*(1+pair.GetSellDelta()), symbol)
			lockedValue, _ := pairStreams.GetAccount().GetLockedAsset(pair.GetPair())
			if (pair.GetUpBound() == 0 || price <= pair.GetUpBound()) && (lockedValue < pair.GetCurrentPositionBalance()) {
				upOrder, err := createOrderInGrid(pairProcessor, binance.SideTypeSell, quantity, price)
				if err != nil {
					return err
				}
				logrus.Debugf("Spots %s: Set Sell order %v on price %v status %v",
					pair.GetPair(), upOrder.OrderID, price, upOrder.Status)
				// Записуємо ордер в грід
				upPrice := grid_types.NewRecord(upOrder.OrderID, price, 0, order.GetPrice(), types.OrderSide(binance.SideTypeSell))
				grid.Set(upPrice)
				order.SetUpPrice(price) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status != binance.OrderStatusTypeNew {
					takerPrice = upPrice
					takerOrder = upOrder
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель нижче
		downPrice, ok := grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
		if ok && downPrice.GetOrderId() == 0 {
			// Створюємо ордер на купівлю
			downOrder, err := createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, order.GetDownPrice())
			if err != nil {
				return err
			}
			downPrice.SetOrderId(downOrder.OrderID)   // Записуємо номер ордера в грід
			downPrice.SetOrderSide(types.SideTypeBuy) // Записуємо сторону ордера в грід
			logrus.Debugf("Spots %s: Set Buy order %v on price %v status %v",
				pair.GetPair(), downOrder.OrderID, order.GetDownPrice(), downOrder.Status)
			if downOrder.Status != binance.OrderStatusTypeNew {
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
				pair,
				pairStreams,
				symbol,
				takerOrder.Side,
				grid,
				takerPrice,
				quantity)
			if err != nil {
				return err
			}
		}
	} else if side == binance.SideTypeBuy {
		// Якшо нижче немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetDownPrice() == 0 {
			// Створюємо ордер на купівлю
			price := roundPrice(order.GetPrice()*(1-pair.GetSellDelta()), symbol)
			if pair.GetLowBound() == 0 || price >= pair.GetLowBound() {
				downOrder, err := createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, price)
				if err != nil {
					return err
				}
				logrus.Debugf("Spots %s: Add Buy order %v on price %v status %v", pair.GetPair(), downOrder.OrderID, price, downOrder.Status)
				// Записуємо ордер в грід
				downPrice := grid_types.NewRecord(downOrder.OrderID, price, order.GetPrice(), 0, types.OrderSide(binance.SideTypeBuy))
				grid.Set(downPrice)
				order.SetDownPrice(price) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status != binance.OrderStatusTypeNew {
					takerPrice = downPrice
					takerOrder = downOrder
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель вище
		upPrice, ok := grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		if ok && upPrice.GetOrderId() == 0 {
			// Створюємо ордер на продаж
			upOrder, err := createOrderInGrid(pairProcessor, binance.SideTypeSell, quantity, order.GetUpPrice())
			if err != nil {
				return err
			}
			if upOrder.Status != binance.OrderStatusTypeNew {
				takerPrice = upPrice
				takerOrder = upOrder
			}
			upPrice.SetOrderId(upOrder.OrderID)      // Записуємо номер ордера в грід
			upPrice.SetOrderSide(types.SideTypeSell) // Записуємо сторону ордера в грід
			logrus.Debugf("Spots %s: Set Sell order %v on price %v status %v",
				pair.GetPair(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status)
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair,
				pairStreams,
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
	grid.Debug("Spots Grid After processOrder", strconv.FormatInt(order.OrderId, 10), pair.GetPair())
	pairProcessor.Debug("Spots Pair After processOrder")
	return
}

func RunSpotGridTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
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
		return err
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
	if pair.GetSellQuantity() == 0 && pair.GetBuyQuantity() == 0 {
		targetValue, err := pairStreams.GetAccount().GetFreeAsset(pair.GetTargetSymbol())
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		pair.SetBuyQuantity(targetValue)
		config.Save()
	}
	// Ініціалізація гріду
	logrus.Debugf("Spot %s: Grid initialized", pair.GetPair())
	grid := grid_types.New()
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		stopEvent <- os.Interrupt
		return fmt.Errorf("spot %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
	}
	symbol, err := func() (res *binance.Symbol, err error) {
		val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.SpotSymbol{Symbol: pair.GetPair()})
		if val == nil {
			return nil, fmt.Errorf("spot %s: Symbol not found", pair.GetPair())
		}
		return val.(*symbol_info.SpotSymbol).GetSpotSymbol()
	}()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Отримання середньої ціни
	price := roundPrice(pair.GetMiddlePrice(), symbol)
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = roundPrice(price, symbol)
	}
	quantity := pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() / price
	minNotional := utils.ConvStrToFloat64(symbol.NotionalFilter().MinNotional)
	if quantity*price < minNotional {
		quantity = utils.RoundToDecimalPlace(minNotional/price, int(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
	}
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, roundPrice(price*(1+pair.GetSellDelta()), symbol), roundPrice(price*(1-pair.GetBuyDelta()), symbol), types.SideTypeNone))
	logrus.Debugf("Spot %s: Set Entry Price order on price %v", pair.GetPair(), price)
	_, err = pairProcessor.CancelAllOrders()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Створюємо ордер на продаж
	sellOrder, err := createOrderInGrid(pairProcessor, binance.SideTypeSell, quantity, roundPrice(price*(1+pair.GetSellDelta()), symbol))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, price*(1+pair.GetSellDelta()), 0, price, types.SideTypeSell))
	logrus.Debugf("Spot %s: Set Sell order on price %v", pair.GetPair(), roundPrice(price*(1+pair.GetSellDelta()), symbol))
	// Створюємо ордер на купівлю
	buyOrder, err := createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, roundPrice(price*(1-pair.GetBuyDelta()), symbol))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, roundPrice(price*(1-pair.GetSellDelta()), symbol), price, 0, types.SideTypeBuy))
	logrus.Debugf("Spot %s: Set Buy order on price %v", pair.GetPair(), roundPrice(price*(1-pair.GetBuyDelta()), symbol))

	// Стартуємо обробку ордерів
	grid.Debug("Spots Grid", "", pair.GetPair())
	logrus.Debugf("Spot %s: Start Order Processing", pair.GetPair())
	mu := &sync.Mutex{}
	for {
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			mu.Lock()
			if config.GetConfigurations().GetReloadConfig() {
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
				balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
				if err != nil {
					stopEvent <- os.Interrupt
					return err
				}
				pair.SetCurrentBalance(balance)
				config.Save()
				quantity = pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() / price
				minNotional := utils.ConvStrToFloat64(symbol.NotionalFilter().MinNotional)
				if quantity*price < minNotional {
					quantity = utils.RoundToDecimalPlace(minNotional/price, int(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
				}
			}
			logrus.Debugf("Spots %s: Order %v on price %v side %v status %s",
				pair.GetPair(),
				event.OrderUpdate.Id,
				event.OrderUpdate.Price,
				event.OrderUpdate.Side,
				event.OrderUpdate.Status)
			// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
			order, ok := grid.Get(&grid_types.Record{Price: utils.ConvStrToFloat64(event.OrderUpdate.Price)}).(*grid_types.Record)
			if !ok {
				logrus.Errorf("Uncorrected order ID: %v", event.OrderUpdate.Id)
				mu.Unlock()
				continue
			}
			err = processOrder(config, pairProcessor, pair, pairStreams, symbol, binance.SideType(event.OrderUpdate.Side), grid, order, quantity)
			if err != nil {
				mu.Unlock()
				pairProcessor.CancelAllOrders()
				stopEvent <- os.Interrupt
				return err
			}
			mu.Unlock()
		case <-time.After(60 * time.Second):
			grid.Debug("Spots Grid", "", pair.GetPair())
		}
	}
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		stopEvent <- os.Interrupt
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		return RunSpotHolding(config, client, degree, limit, pair, stopEvent, updateTime, debug)

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		return RunSpotScalping(config, client, degree, limit, pair, stopEvent, updateTime, debug)

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		return RunSpotTrading(config, client, degree, limit, pair, stopEvent, updateTime, debug)

		// Відпрацьовуємо Grid стратегію
	} else if pair.GetStrategy() == pairs_types.GridStrategyType {
		return RunSpotGridTrading(config, client, pair, stopEvent)

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		stopEvent <- os.Interrupt
		return fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
	}
}
