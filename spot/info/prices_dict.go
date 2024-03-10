package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
)

var (
	PricesMap      = make(map[SymbolName]SymbolPrice)
	mu_prices_dict sync.Mutex
)

func InitPricesMap(client *binance.Client, symbolname string) (err error) {
	res, err :=
		client.NewListPricesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_prices_dict.Lock()
	defer mu_prices_dict.Unlock()
	for _, price := range res {
		PricesMap[SymbolName(price.Symbol)] = SymbolPrice(utils.ConvStrToFloat64(price.Price))
	}
	return nil
}

func GetPricesMap() map[SymbolName]SymbolPrice {
	mu_prices_dict.Lock()
	defer mu_prices_dict.Unlock()
	return PricesMap
}

func GetPrice(symbolname SymbolName) SymbolPrice {
	mu_prices_dict.Lock()
	defer mu_prices_dict.Unlock()
	return PricesMap[symbolname]
}
