package services

import (
	"context"

	"github.com/adshao/go-binance/v2"
)

func GetExchangeInfo(client *binance.Client) (*binance.ExchangeInfo, error) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	return exchangeInfo, err
}

func GetOrderTypes(exchangeInfo *binance.ExchangeInfo, symbolname string) []binance.OrderType {
	res := make([]binance.OrderType, 0)
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, orderType := range info.OrderTypes {
				res = append(res, binance.OrderType(orderType))
			}
		}
	}
	return res
}

func GetPermissions(exchangeInfo *binance.ExchangeInfo, symbolname string) []string {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			return info.Permissions
		}
	}
	return nil
}

func GetFilters(exchangeInfo *binance.ExchangeInfo, symbolname string) []map[string]interface{} { //[]binance.SymbolFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			return info.Filters
		}
	}
	return nil
}
