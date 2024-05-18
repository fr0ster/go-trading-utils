package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
)

func GetAccountInfoGuard(
	account *futures_account.Account,
	source chan *futures.WsUserDataEvent) (out chan *futures.WsUserDataEvent) {
	out = make(chan *futures.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == futures.UserDataEventTypeAccountUpdate {
				if account.UpdateTime < event.Time {
					account.Lock()
					logrus.Debugf("Account update Reason: %s", event.AccountUpdate.Reason)
					for _, item := range event.AccountUpdate.Balances {
						val, _ := futures_account.Futures2AccountAsset(item)
						account.AssetUpdate(val)
					}
					for _, item := range event.AccountUpdate.Positions {
						val, _ := futures_account.Futures2AccountPosition(item)
						account.PositionUpdate(val)
					}
					account.Unlock()
					out <- event
				}
			}
		}
	}()
	return
}
