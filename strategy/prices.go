package strategy

import (
	"context"
	"errors"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
)

func GetLimitPricesDumpWay(data types.Config, client *binance.Client) (string, float64, string, string, string, string, string, string) {
	balance := data.Balance
	symbolname := data.Symbol

	priceF, _, err := GetMarketPrice(client, string(symbolname))
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

func BidOrAsk(data types.Config, bookTickers *bookticker_types.BookTickerBTree, client *binance.Client, side string) (price, targetPrice, targetQuantity, stopPriceSL, priceSL, stopPriceTP, priceTP, trailingDelta string) {
	// При налаштуванні лімітного ордера на продаж, ви, як правило, орієнтуєтесь на ціну bid.
	// Ціна bid - це найвища ціна, яку покупець готовий заплатити за актив.
	// Коли ви продаете, ви хочете отримати найвищу можливу ціну,
	// тому ви встановлюєте свій лімітний ордер на продаж на рівні ціни bid або вище.
	// Ціна ask, з іншого боку, - це найнижча ціна, за яку продавець готовий продати актив.
	// Коли ви купуєте, ви хочете заплатити найнижчу можливу ціну,
	// тому ви встановлюєте свій лімітний ордер на купівлю на рівні ціни ask або нижче.

	balance := data.Balance
	symbolname := data.Symbol
	targetPriceF := 0.0

	bookTicker, err := bookticker_types.Binance2BookTicker(bookTickers.Get(symbolname))
	if err != nil {
		utils.HandleErr(errors.New("BookTicker not found"))
	}
	if bookTicker == nil {
		utils.HandleErr(errors.New("BookTicker not found"))
	} else {
		if side == "BUY" {
			targetPriceF = float64(bookTicker.AskPrice) * 0.9
		} else {
			targetPriceF = float64(bookTicker.BidPrice) * 1.1
		}
	}

	targetQuantity = utils.ConvFloat64ToStr(balance*0.01/targetPriceF, 8)
	targetPrice = utils.ConvFloat64ToStr(targetPriceF, 8)

	price = utils.ConvFloat64ToStrDefault(targetPriceF)
	trailingDelta = "100"
	stopPriceSL = utils.ConvFloat64ToStrDefault(targetPriceF * 0.95)
	priceSL = utils.ConvFloat64ToStrDefault(targetPriceF * 0.90)
	stopPriceTP = utils.ConvFloat64ToStrDefault(targetPriceF * 1.05)
	priceTP = utils.ConvFloat64ToStrDefault(targetPriceF * 1.10)

	return price, targetPrice, targetQuantity, stopPriceSL, priceSL, stopPriceTP, priceTP, trailingDelta
}

func GetMarketPrice(client *binance.Client, symbol string) (float64, string, error) {
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		return 0, "", err
	}
	marketPrice := prices[0]
	return utils.ConvStrToFloat64(marketPrice.Price), marketPrice.Price, err
}
