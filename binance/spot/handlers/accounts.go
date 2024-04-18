package handlers

import (
	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
)

func GetChangingOfAccountInfoGuard(
	account *spot_account.Account,
	source chan *binance.WsUserDataEvent) (out chan *binance.WsUserDataEvent) {
	out = make(chan *binance.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == binance.UserDataEventTypeOutboundAccountPosition {
				if account.AccountUpdateTime < event.AccountUpdate.AccountUpdateTime {
					account.Lock() // Locking the balances
					for _, item := range event.AccountUpdate.WsAccountUpdates {
						account.AssetUpdate(binance.Balance(item))
					}
					account.Unlock() // Unlocking the balances
				}
			}
		}
	}()
	return
}
