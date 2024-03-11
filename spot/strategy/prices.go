package strategy

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/services"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetLimitPricesDumpWay(data utils.DataRecord, client *binance.Client) (string, float64, string, string, string, string, string, string) {
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
	return price, targetPrice, targetQuantity, stopPriceSL, priceSL, stopPriceTP, priceTP, trailingDelta
}

func BidOrAsk(client *binance.Client, symbol string, price string, quantity string, side binance.SideType) {
	// При налаштуванні лімітного ордера на продаж, ви, як правило, орієнтуєтесь на ціну bid.
	// Ціна bid - це найвища ціна, яку покупець готовий заплатити за актив.
	// Коли ви продаете, ви хочете отримати найвищу можливу ціну,
	// тому ви встановлюєте свій лімітний ордер на продаж на рівні ціни bid або вище.
	// Ціна ask, з іншого боку, - це найнижча ціна, за яку продавець готовий продати актив.
	// Коли ви купуєте, ви хочете заплатити найнижчу можливу ціну,
	// тому ви встановлюєте свій лімітний ордер на купівлю на рівні ціни ask або нижче.
}
