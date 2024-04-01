package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBalancesUpdateGuard(bt *balances_types.BalanceBTree, source chan *futures.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.Balances {
				balanceUpdate := &balances_types.BalanceItemType{
					Asset:  balances_types.AssetType(item.Asset),
					Free:   utils.ConvStrToFloat64(item.Balance),
					Locked: utils.ConvStrToFloat64(item.Balance) - utils.ConvStrToFloat64(item.ChangeBalance),
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
