package spot_signals

import (
	"errors"
	"math"
	_ "net/http/pprof"

	"os"

	"github.com/google/btree"
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
