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
			for _, item := range event.AccountUpdate.WsAccountUpdates {
				balanceUpdate := &balances_types.BalanceItemType{
					Asset:  balances_types.AssetType(item.Asset),
					Free:   utils.ConvStrToFloat64(item.Free),
					Locked: utils.ConvStrToFloat64(item.Locked),
				}
				bt.Lock()
				bt.SetItem(balanceUpdate)
				bt.Unlock()
			}
			out <- true
		}
	}()
	return
}
