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
						if val := account.GetAssets().Get(&futures_account.Asset{Asset: item.Asset}); val != nil {
							val.(*futures_account.Asset).WalletBalance = item.Balance
							val.(*futures_account.Asset).CrossWalletBalance = item.CrossWalletBalance
						}
						account.SetAssetsUpd(&item)
					}
					for _, item := range event.AccountUpdate.Positions {
						if val := account.GetPositions().Get(&futures_account.Position{Symbol: item.Symbol}); val != nil {
							val.(*futures_account.Position).PositionSide = item.Side
							val.(*futures_account.Position).PositionAmt = item.Amount
							val.(*futures_account.Position).EntryPrice = item.EntryPrice
							val.(*futures_account.Position).UnrealizedProfit = item.UnrealizedPnL
						}
						account.SetPositionsUpd(&item)
					}
					account.Unlock()
					out <- event
				}
			}
		}
	}()
	return
}
