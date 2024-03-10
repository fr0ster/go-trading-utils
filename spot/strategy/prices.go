package strategy

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/services"
	"github.com/fr0ster/go-binance-utils/spot/utils"
)

func GetLimitPricesDumpWay(data utils.DataRecord, client *binance.Client) (string, string, string, string, string, string, string) {
	balance := data.Balance
	symbolname := data.Symbol

	priceF, _, err := services.GetMarketPrice(client, string(symbolname))
	if err != nil {
		utils.HandleErr(err)
	}
	targetQuantity := utils.ConvFloat64ToStr(balance*0.01/priceF, 8)

	targetPrice := priceF * 0.9
	stopPriceSLF := targetPrice * 0.95
	priceSLF := targetPrice * 0.90
	stopPriceTPF := targetPrice * 1.05
	priceTPF := targetPrice * 1.10

	price := utils.ConvFloat64ToStrDefault(priceF)
	trailingDelta := "100"
	stopPriceSL := utils.ConvFloat64ToStrDefault(stopPriceSLF)
	priceSL := utils.ConvFloat64ToStrDefault(priceSLF)
	stopPriceTP := utils.ConvFloat64ToStrDefault(stopPriceTPF)
	priceTP := utils.ConvFloat64ToStrDefault(priceTPF)
	return price, targetQuantity, stopPriceSL, priceSL, stopPriceTP, priceTP, trailingDelta
}
