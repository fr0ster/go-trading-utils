package services

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
)

func GetMarketPrice(client *binance.Client, symbol string) (float64, error) {
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	marketPrice := prices[0]
	return utils.ConvStrToFloat64(marketPrice.Price), err
}
