package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/balances"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetBalancesUpdateGuard(bt *balances.BalanceBTree, source chan *futures.WsUserDataEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			for _, item := range event.AccountUpdate.Balances {
				balanceUpdate := balances.BalanceItemType{
					Asset:  balances.AssetType(item.Asset),
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
