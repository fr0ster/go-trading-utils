package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetBalanceTreeUpdateHandler() (wsHandler binance.WsUserDataHandler, accountEventChan chan bool) {
	accountEventChan = make(chan bool)
	wsHandler = func(event *binance.WsUserDataEvent) {
		for _, item := range event.AccountUpdate.WsAccountUpdates {
			accountUpdate := markets.BalanceItemType{
				Asset:  item.Asset,
				Free:   utils.ConvStrToFloat64(item.Free),
				Locked: utils.ConvStrToFloat64(item.Locked),
			}
			markets.GetBalancesTree().ReplaceOrInsert(accountUpdate)
		}
		accountEventChan <- true
	}
	return
}
