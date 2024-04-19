package handlers

import (
	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
)

func GetChangingOfAccountInfoGuard(
	account *futures_account.Account,
	source chan *futures.WsUserDataEvent) (out chan *futures.WsUserDataEvent) {
	out = make(chan *futures.WsUserDataEvent, 1)
	go func() {
		for {
			event := <-source
			if event.Event == futures.UserDataEventTypeAccountUpdate {
				if account.AccountUpdateTime < event.Time {
					account.Lock()
					for _, val := range event.AccountUpdate.Balances {
						account.AssetUpdate(&futures_account.Asset{Asset: val.Asset, WalletBalance: val.Balance, CrossWalletBalance: val.CrossWalletBalance})
					}
					account.Unlock()
					out <- event
				}
			}
		}
	}()
	return
}
