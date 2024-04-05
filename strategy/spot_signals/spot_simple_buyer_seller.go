package spot_signals

import (
	"context"
	"errors"
	"math"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/info"
	spot_bookticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	utils "github.com/fr0ster/go-trading-utils/utils"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

const (
	errorMsg = "Error: %v"
)

func LimitRead(degree int, symbols []string, client *binance.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	exchangeInfo := exchange_types.NewExchangeInfo()
	spot_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}

func StartPairStreams(
	symbol string,
	bookTicker *bookTicker_types.BookTickerBTree,
	depth *depth_types.Depth) (
	depthEvent chan bool,
	bookTickerEvent chan bool) {
	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(symbol, 1)
	bookTickerStream.Start()

	bookTickerEvent = spot_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	// Запускаємо потік для отримання оновлення стакана
	depthStream := spot_streams.NewDepthStream(symbol, true, 1)
	depthStream.Start()

	depthEvent = spot_handlers.GetDepthsUpdateGuard(depth, depthStream.DataChannel)

	return
}

func StartGlobalStreams(
	client *binance.Client,
	stop chan os.Signal,
	balances *balances_types.BalanceBTree) (
	userDataStream4Balance *spot_streams.UserDataStream,
	balanceEvent chan bool,
	userDataStream4Order *spot_streams.UserDataStream,
	orderStatusEvent chan *binance.WsUserDataEvent) {
	// Запускаємо потік для отримання wsUserDataEvent
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		logrus.Errorf(errorMsg, err)
		stop <- os.Interrupt
		return
	}
	userDataStream4Order = spot_streams.NewUserDataStream(listenKey, 1)
	userDataStream4Order.Start()

	orderStatuses := []binance.OrderStatusType{
		binance.OrderStatusTypeFilled,
		binance.OrderStatusTypePartiallyFilled,
	}

	orderStatusEvent = spot_handlers.GetChangingOfOrdersGuard(
		userDataStream4Order.DataChannel,
		binance.UserDataEventTypeExecutionReport,
		orderStatuses)

	userDataStream4Balance = spot_streams.NewUserDataStream(listenKey, 1)
	userDataStream4Balance.Start()

	// Запускаємо потік для отримання оновлення балансу
	balanceEvent = spot_handlers.GetBalancesUpdateGuard(balances, userDataStream4Balance.DataChannel)

	return
}

func RestUpdate(
	client *binance.Client,
	stop chan os.Signal,
	pair *config_interfaces.Pairs,
	depth *depth_types.Depth,
	limit int,
	bookTicker *bookTicker_types.BookTickerBTree,
	updateTime time.Duration) {
	go func() {
		for {
			select {
			case <-stop:
				// Якщо отримано сигнал з каналу stop, вийти з циклу
				return
			default:
				err := spot_depth.SpotDepthInit(depth, client, limit)
				if err != nil {
					logrus.Errorf(errorMsg, err)
					stop <- os.Interrupt
					return
				}

				time.Sleep(1 * time.Second)

				err = spot_bookticker.Init(bookTicker, (*pair).GetPair(), client)
				if err != nil {
					logrus.Errorf(errorMsg, err)
					stop <- os.Interrupt
					return
				}

				time.Sleep(updateTime)
			}
		}
	}()
}

func Spot_depth_buy_sell_signals(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (buyEvent chan *depth_types.DepthItemType, sellEvent chan *depth_types.DepthItemType) {
	var boundAsk float64
	var boundBid float64
	buyEvent = make(chan *depth_types.DepthItemType, 1)
	sellEvent = make(chan *depth_types.DepthItemType, 1)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				getBaseBalance := func(pair *config_interfaces.Pairs) (
					baseBalance float64,
					err error) {
					baseBalance, err = account.GetAsset((*pair).GetBaseSymbol())
					return
				}
				getTargetBalance := func(pair *config_interfaces.Pairs) (
					targetBalance float64,
					err error) {
					targetBalance, err = account.GetAsset((*pair).GetTargetSymbol())
					return
				}
				baseBalance, err := getBaseBalance(pair)
				if err != nil {
					logrus.Warnf("Can't get %s balance: %v", (*pair).GetTargetSymbol(), err)
					continue
				}
				targetBalance, err := getTargetBalance(pair)
				if err != nil {
					logrus.Warnf("Can't get %s balance: %v", (*pair).GetTargetSymbol(), err)

					continue
				}
				limitBalance := (*pair).GetLimit()

				getAskAndBid := func(depths *depth_types.Depth) (ask float64, bid float64, err error) {
					getPrice := func(val btree.Item) float64 {
						if val == nil {
							err = errors.New("value is nil")
						}
						return val.(*depth_types.DepthItemType).Price
					}
					ask = getPrice(depths.GetAsks().Min())
					bid = getPrice(depths.GetBids().Max())
					return
				}

				ask, bid, err := getAskAndBid(depths)
				if err != nil {
					logrus.Warnf("Can't get ask and bid: %v", err)
					continue
				}

				getBound := func(pair *config_interfaces.Pairs) (boundAsk float64, boundBid float64, err error) {
					if boundAsk == ask*(1+(*pair).GetBuyDelta()) &&
						boundBid == bid*(1-(*pair).GetSellDelta()) {
						err = errors.New("bounds are the same")
					} else {
						boundAsk = ask * (1 + (*pair).GetBuyDelta())
						logrus.Debugf("Ask bound: %f", boundAsk)
						boundBid = bid * (1 - (*pair).GetSellDelta())
						logrus.Debugf("Bid bound: %f", boundBid)
					}
					return
				}
				boundAsk, boundBid, err = getBound(pair)
				if err != nil {
					logrus.Warnf("Can't get bounds: %v", err)
					continue
				}
				// Value for BUY and SELL transactions
				limitValue := (*pair).GetLimitOnTransaction() * limitBalance // Value for one transaction

				// SELL Quantity for one transaction
				sellQuantity := limitValue / bid // Quantity for one SELL transaction
				if sellQuantity > targetBalance {
					sellQuantity = targetBalance // Quantity for one SELL transaction if it's more than available
				}

				// Correct value for BUY transaction
				if limitValue > math.Min(limitBalance, baseBalance) {
					limitValue = math.Min(limitBalance, baseBalance)
				}
				// BUY Quantity for one transaction
				buyQuantity := limitValue / boundAsk
				// If quantity for one BUY transaction is less than available
				if buyQuantity*boundAsk < baseBalance &&
					// And middle price is higher than low bound price
					((*pair).GetMiddlePrice() == 0 || (*pair).GetMiddlePrice() >= boundAsk) {
					logrus.Infof("Middle price %f is higher than high bound price %f, BUY!!!", (*pair).GetMiddlePrice(), boundAsk)
					buyEvent <- &depth_types.DepthItemType{
						Price:    boundAsk,
						Quantity: buyQuantity}
					// If quantity for one SELL transaction is less than available
				} else if sellQuantity <= targetBalance &&
					// And middle price is lower than low bound price
					(*pair).GetMiddlePrice() <= boundBid {
					logrus.Infof("Middle price %f is lower than low bound price %f, SELL!!!", (*pair).GetMiddlePrice(), boundBid)
					sellEvent <- &depth_types.DepthItemType{
						Price:    boundBid,
						Quantity: sellQuantity}
				} else {
					targetAsk := (*pair).GetMiddlePrice() * (1 - (*pair).GetBuyDelta())
					targetBid := (*pair).GetMiddlePrice() * (1 + (*pair).GetSellDelta())
					if baseBalance < limitBalance {
						logrus.Infof("Now ask is %f, bid is %f", ask, bid)
						logrus.Infof("Waiting for bid increase to %f", targetBid)
					} else {
						logrus.Infof("Now ask is %f, bid is %f", ask, bid)
						logrus.Infof("Waiting for ask decrease to %f or bid increase to %f", targetAsk, targetBid)
					}
				}
				logrus.Infof("Current profit: %f", (*pair).GetProfit(bid))
				logrus.Infof("Predicable profit: %f", (*pair).GetProfit((*pair).GetMiddlePrice()*(1+(*pair).GetSellDelta())))
				logrus.Infof("Middle price: %f, available USDT: %f, Bid: %f", (*pair).GetMiddlePrice(), baseBalance, bid)
			}
		}
	}()
	return
}

func Process(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	buyEvent chan *depth_types.DepthItemType,
	sellEvent chan *depth_types.DepthItemType,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent) {
	var (
		startBuyOrderEvent  = make(chan *binance.CreateOrderResponse)
		startSellOrderEvent = make(chan *binance.CreateOrderResponse)
		quantityRound       = int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).LotSizeFilter().StepSize)))
		priceRound          = int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).PriceFilter().TickSize)))
	)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case params := <-buyEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				order, err :=
					client.NewCreateOrderService().
						Symbol(string(binance.SymbolType((*pair).GetPair()))).
						Type(binance.OrderTypeLimit).
						Side(binance.SideTypeBuy).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
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
					(*pair).SetBuyQuantity((*pair).GetBuyQuantity() + utils.ConvStrToFloat64(order.ExecutedQuantity))
					(*pair).SetBuyValue((*pair).GetBuyValue() + utils.ConvStrToFloat64(order.ExecutedQuantity)*utils.ConvStrToFloat64(order.Price))
					config.Save()
				}
			case params := <-sellEvent:
				if minuteOrderLimit.Limit == 0 || dayOrderLimit.Limit == 0 || minuteRawRequestLimit.Limit == 0 {
					logrus.Warn("Order limits has been out!!!, waiting for update...")
					continue
				}
				order, err :=
					client.NewCreateOrderService().
						Symbol(string(binance.SymbolType((*pair).GetPair()))).
						Type(binance.OrderTypeLimit).
						Side(binance.SideTypeSell).
						Quantity(utils.ConvFloat64ToStr(params.Quantity, quantityRound)).
						Price(utils.ConvFloat64ToStr(params.Price, priceRound)).
						TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
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
					(*pair).SetSellQuantity((*pair).GetSellQuantity() + utils.ConvStrToFloat64(order.ExecutedQuantity))
					(*pair).SetSellValue((*pair).GetSellValue() + utils.ConvStrToFloat64(order.ExecutedQuantity)*utils.ConvStrToFloat64(order.Price))
					config.Save()
				}
			}
		}
	}()
	go func() {
		for {
			select {
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
								config.Save()
								break
							}
						}
					}
				}
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

func GetPrice(client *binance.Client, symbol string) (float64, error) {
	price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	var (
		depth      *depth_types.Depth
		bookTicker *bookTicker_types.BookTickerBTree
	)

	depth = depth_types.NewDepth(degree, (*pair).GetPair())

	bookTicker = bookTicker_types.New(degree)

	_, bookTickerEvent := StartPairStreams((*pair).GetPair(), bookTicker, depth)

	RestUpdate(client, stopEvent, pair, depth, limit, bookTicker, updateTime)

	price, err := GetPrice(client, (*pair).GetPair())
	if err != nil {
		logrus.Errorf("Can't get price: %v", err)
		stopEvent <- os.Interrupt
		return
	}

	if (*pair).GetBuyQuantity() == 0 && (*pair).GetSellQuantity() == 0 {
		targetFree, err := account.GetAsset((*pair).GetTargetSymbol())
		if err != nil {
			logrus.Errorf("Can't get %s asset: %v", (*pair).GetTargetSymbol(), err)
			stopEvent <- os.Interrupt
			return
		}
		(*pair).SetBuyQuantity(targetFree)
		(*pair).SetBuyValue(targetFree * price)
	}
	config.Save()

	buyEvent, sellEvent := Spot_depth_buy_sell_signals(account, depth, pair, stopEvent, bookTickerEvent)

	Process(
		config, client, pair, pairInfo,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, sellEvent, stopEvent, orderStatusEvent)

	go func() {
		for {
			baseBalance, err := account.GetAsset((*pair).GetBaseSymbol())
			if err != nil {
				logrus.Errorf("Can't get %s asset: %v", (*pair).GetBaseSymbol(), err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				return
			default:
				if val := (*pair).GetMiddlePrice(); val != 0 {
					logrus.Infof("Middle %s price: %f, available USDT: %f, Price: %f",
						(*pair).GetPair(), val, baseBalance, price)
				}
			}
			time.Sleep(updateTime)
		}
	}()
}
