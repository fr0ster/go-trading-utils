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

func PrintFiltersInfo(client *binance.Client, symbolname string) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching exchange info: %v", err)
	}

	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filters := range info.Filters {
				for _, filter := range filters {
					if filter == "PRICE_FILTER" {
						fmt.Printf("Filter: %s\n", filter)
					}
				}
			}
		}
	}
}
