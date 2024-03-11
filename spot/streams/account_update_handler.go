package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetBalanceTreeUpdateHandler() (wsHandler binance.WsDepthHandler, accountEventChan chan bool) {
	accountEventChan = make(chan bool)
	wsHandler = func(event *binance.WsDepthEvent) {
		accountUpdate := info.BalanceItemType{
			Asset:  event.Symbol,
			Free:   utils.ConvStrToFloat64(event.Bids[0].Price),
			Locked: utils.ConvStrToFloat64(event.Asks[0].Price),
		}
		info.SetBalanceTreeItem(accountUpdate)
		accountEventChan <- true
	}
	return
}
