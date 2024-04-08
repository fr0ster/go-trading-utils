package spot_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

func Spot_depth_buy_sell_signals(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (buyEvent chan *depth_types.DepthItemType, sellEvent chan *depth_types.DepthItemType) {
	buyEvent = make(chan *depth_types.DepthItemType, 1)
	sellEvent = make(chan *depth_types.DepthItemType, 1)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				baseBalance, // Кількість базової валюти
					targetBalance, // Кількість торгової валюти
					_,             // limitBalance, // Ліміт на купівлю повний у одиницях базової валюти
					_,             // LimitInputIntoPosition, // Ліміт на вхід в позицію, відсоток від балансу базової валюти
					_,             //LimitInPosition, // Ліміт базової валюти на одну позицію купівлі або продажу
					ask,           // Ціна купівлі
					bid,           // Ціна продажу
					boundAsk,      // Верхня межа ціни купівлі
					boundBid,      // Нижня межа ціни продажу
					_,             //transactionValue, // Ліміт на купівлю на одну позицію купівлі або продажу
					sellQuantity,  // Кількість торгової валюти для продажу
					buyQuantity,   // Кількість торгової валюти для купівлі
					err := getData4Analysis(account, depths, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				if ((*pair).GetMiddlePrice() == 0 || // Якшо середня ціна купівли котирувальної валюти дорівнює нулю
					(*pair).GetMiddlePrice() >= boundAsk) && // Та середня ціна купівли котирувальної валюти більша або дорівнює верхній межі ціни купівли
					buyQuantity > 0 { // Та кількість цільової валюти для купівлі більша за нуль
					logrus.Infof("Middle price %f is higher than high bound price %f, BUY!!!", (*pair).GetMiddlePrice(), boundAsk)
					buyEvent <- &depth_types.DepthItemType{
						Price:    boundAsk,
						Quantity: buyQuantity}
				} else if (*pair).GetMiddlePrice() <= boundBid && // Якшо середня ціна купівли котирувальної валюти менша або дорівнює нижній межі ціни продажу
					sellQuantity > 0 { // Та кількість цільової валюти для продажу більша за нуль
					logrus.Infof("Middle price %f is lower than low bound price %f, SELL!!!", (*pair).GetMiddlePrice(), boundBid)
					sellEvent <- &depth_types.DepthItemType{
						Price:    boundBid,
						Quantity: sellQuantity}
				}
				logrus.Infof("Now ask is %f, bid is %f", ask, bid)
				logrus.Infof("Current Ask bound: %f, Bid bound: %f", boundAsk, boundBid)
				logrus.Infof("Middle price: %f, available USDT: %f, available %s: %f",
					(*pair).GetMiddlePrice(), baseBalance, (*pair).GetTargetSymbol(), targetBalance)
				logrus.Infof("Current profit: %f", (*pair).GetProfit(bid))
				logrus.Infof("Predicable profit: %f", (*pair).GetProfit((*pair).GetMiddlePrice()*(1+(*pair).GetSellDelta())))
				time.Sleep(5 * time.Second)
			}
		}
	}()
	return
}

func getData4Analysis(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs) (
	baseBalance float64, // Кількість базової валюти
	targetBalance float64, // Кількість торгової валюти
	LimitInputIntoPosition float64, // Ліміт на вхід в позицію, відсоток від балансу базової валюти
	LimitInPosition float64, // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction float64, // Ліміт на транзакцію, відсоток від ліміту на позицію
	ask float64, // Ціна купівлі
	bid float64, // Ціна продажу
	boundAsk float64, // Верхня межа ціни купівлі
	boundBid float64, // Нижня межа ціни продажу
	transactionValue float64, // Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	sellQuantity float64, // Кількість торгової валюти для продажу
	buyQuantity float64, // Кількість торгової валюти для купівлі
	err error) {
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
	baseBalance, err = getBaseBalance(pair)
	if err != nil {
		logrus.Warnf("Can't get %s balance: %v", (*pair).GetTargetSymbol(), err)
		return
	}
	targetBalance, err = getTargetBalance(pair)
	if err != nil {
		logrus.Warnf("Can't get %s balance: %v", (*pair).GetTargetSymbol(), err)
		return
	}

	// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
	// поки не наберемо цей ліміт, не можемо перейти до режиму спекуляціі
	// Режим входу - накопичуємо цільовий токен
	// Режим спекуляції - купуємо/продаемо цільовий токен за базовий
	// Режим виходу - продаемо цільовий токен
	LimitInputIntoPosition = (*pair).GetLimitInputIntoPosition()
	// Ліміт на позицію, відсоток від балансу базової валюти
	LimitInPosition = (*pair).GetLimitInPosition()
	// Ліміт на транзакцію, відсоток від ліміту на позицію
	LimitOnTransaction = (*pair).GetLimitOnTransaction()
	// Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	transactionValue = LimitOnTransaction * LimitInPosition * baseBalance

	ask, bid, err = GetAskAndBid(depths)
	if err != nil {
		logrus.Warnf("Can't get ask and bid: %v", err)
		return
	}

	boundAsk, boundBid, err = GetBound(pair)
	if err != nil {
		logrus.Warnf("Can't get bounds: %v", err)
		return
	}

	// Кількість торгової валюти для продажу
	sellQuantity = transactionValue / bid
	if sellQuantity > targetBalance {
		sellQuantity = targetBalance // Якщо кількість торгової валюти для продажу більша за доступну, то продаємо доступну
	}

	// Кількість торгової валюти для купівлі
	buyQuantity = transactionValue / boundAsk
	return
}

func BuyOrSellSignal(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (buyEvent chan *depth_types.DepthItemType, sellEvent chan *depth_types.DepthItemType) {
	buyEvent = make(chan *depth_types.DepthItemType, 1)
	sellEvent = make(chan *depth_types.DepthItemType, 1)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				_, _, _, _, _,
					ask,      // Ціна купівлі
					bid,      // Ціна продажу
					boundAsk, // Верхня межа ціни купівлі
					boundBid, // Нижня межа ціни продажу
					_,
					sellQuantity, // Кількість торгової валюти для продажу
					buyQuantity,  // Кількість торгової валюти для купівлі
					err := getData4Analysis(account, depths, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Середня ціна купівли котирувальної валюти дорівнює нулю або більша за верхню межу ціни купівли
				if (*pair).GetMiddlePrice() >= boundAsk &&
					buyQuantity > 0 { // Та кількість цільової валюти для купівлі більша за нуль
					logrus.Infof("Middle price %f is higher than high bound price %f, BUY!!!", (*pair).GetMiddlePrice(), boundAsk)
					buyEvent <- &depth_types.DepthItemType{
						Price:    boundAsk,
						Quantity: buyQuantity}
					// Середня ціна купівли котирувальної валюти менша або дорівнює нижній межі ціни продажу
				} else if (*pair).GetMiddlePrice() <= boundBid &&
					sellQuantity > 0 { // Та кількість цільової валюти для продажу більша за нуль
					logrus.Infof("Middle price %f is lower than low bound price %f, SELL!!!", (*pair).GetMiddlePrice(), boundBid)
					sellEvent <- &depth_types.DepthItemType{
						Price:    boundBid,
						Quantity: sellQuantity}
					// Чекаємо на зміну ціни
				} else {
					if buyQuantity == 0 || sellQuantity == 0 {
						logrus.Info("Wait for buy signal")
						logrus.Infof("Now ask is %f, bid is %f", ask, bid)
						logrus.Infof("Waiting for ask decrease to %f", boundAsk)
					} else if (*pair).GetMiddlePrice() > boundBid && (*pair).GetMiddlePrice() < boundAsk {
						logrus.Infof("Now ask is %f, bid is %f", ask, bid)
						logrus.Infof("Waiting for ask decrease to %f or bid increase to %f", boundAsk, boundBid)
					}
				}
			}
		}
	}()
	return
}

func InPositionSignal(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	timeFrame time.Duration,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (
	collectionEvent chan *depth_types.DepthItemType, // Накопичуемо цільову валюту
	positionEvent chan *depth_types.DepthItemType) { // Переходимо в режим спекуляції
	collectionEvent = make(chan *depth_types.DepthItemType, 1)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
			case <-time.After(timeFrame): // Або просто чекаємо якийсь час
			default:
				continue
			}
			baseBalance, // Кількість базової валюти
				targetBalance,          // Кількість торгової валюти
				LimitInputIntoPosition, // Ліміт на вхід в позицію, відсоток від балансу базової валюти
				LimitInPosition,        // Ліміт на позицію, відсоток від балансу базової валюти
				_,                      // LimitOnTransaction,     // Ліміт на транзакцію, відсоток від ліміту на позицію
				ask,                    // Ціна купівлі
				bid,                    // Ціна продажу
				boundAsk,               // Верхня межа ціни купівлі
				_,                      // Нижня межа ціни продажу
				_,                      // limitValue, // Ліміт на купівлю на одну позицію купівлі або продажу
				_,                      // Кількість торгової валюти для продажу
				buyQuantity,            // Кількість торгової валюти для купівлі
				err := getData4Analysis(account, depths, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
			if targetBalance*boundAsk > baseBalance*LimitInputIntoPosition*LimitInPosition {
				positionEvent <- &depth_types.DepthItemType{
					Price:    boundAsk,
					Quantity: buyQuantity}
				return
				// Якшо вартість цільової валюти менша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - накопичуємо
			} else if targetBalance*boundAsk < baseBalance*LimitInputIntoPosition*LimitInPosition {
				logrus.Infof("Middle price %f is higher than high bound price %f, BUY!!!", (*pair).GetMiddlePrice(), boundAsk)
				collectionEvent <- &depth_types.DepthItemType{
					Price:    boundAsk,
					Quantity: buyQuantity}
			} else {
				targetAsk := (*pair).GetMiddlePrice() * (1 - (*pair).GetBuyDelta())
				if ask > targetAsk {
					logrus.Infof("Now ask is %f, bid is %f", ask, bid)
					logrus.Infof("Waiting for ask decrease to %f", targetAsk)
				}
			}
			logrus.Infof("Current profit: %f", (*pair).GetProfit(bid))
			logrus.Infof("Predicable profit: %f", (*pair).GetProfit((*pair).GetMiddlePrice()*(1+(*pair).GetSellDelta())))
			logrus.Infof("Middle price: %f, available USDT: %f, Bid: %f", (*pair).GetMiddlePrice(), baseBalance, bid)
		}
	}()
	return
}
