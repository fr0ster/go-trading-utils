package services

import (
	"context"
	"fmt"
	"log"

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

func PrintFiltersInfo(client *binance.Client, symbolname string) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching exchange info: %v", err)
	}

	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filters := range info.Filters {
				for _, filter := range filters {
					fmt.Printf("Filter: %s\n", filter)
				}
			}
		}
	}
}
