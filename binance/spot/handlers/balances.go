package handlers

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/balances"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBalancesUpdateGuard(bt *balances.BalanceBTree, source chan *binance.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.WsAccountUpdates {
				balanceUpdate := balances.BalanceItemType{
					Asset:  balances.AssetType(item.Asset),
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
