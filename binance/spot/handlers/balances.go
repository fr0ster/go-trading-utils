package handlers

import (
	"github.com/adshao/go-binance/v2"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBalancesUpdateGuard(bt *balances_types.BalanceBTree, source chan *binance.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			if event.Event == binance.UserDataEventTypeOutboundAccountPosition {
				bt.Lock() // Locking the balances
				for _, item := range event.AccountUpdate.WsAccountUpdates {
					balanceUpdate := &balances_types.BalanceItemType{
						Asset:  item.Asset,
						Free:   utils.ConvStrToFloat64(item.Free),
						Locked: utils.ConvStrToFloat64(item.Locked),
					}
					bt.SetItem(balanceUpdate)
				}
				bt.Unlock() // Unlocking the balances
			}
			out <- true
		}
	}()
	return
}
