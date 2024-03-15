package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
	symbols_info "github.com/fr0ster/go-binance-utils/spot/info/symbols"
	symbol_info "github.com/fr0ster/go-binance-utils/spot/info/symbols/symbol"
)

type ExchangeInfo struct {
	exchangeInfo *binance.ExchangeInfo
	Symbols      *symbols_info.Symbols
}

func GetExchangeInfo(client *binance.Client) (*ExchangeInfo, error) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	symbols := symbols_info.NewSymbols(2, exchangeInfo.Symbols)
	return &ExchangeInfo{exchangeInfo, symbols}, nil
}

func (exchangeInfo *ExchangeInfo) GetSymbol(symbol string) *symbol_info.Symbol {
	return exchangeInfo.Symbols.GetSymbol(symbol)
}

// func (exchangeInfo *ExchangeInfo) GetOrderTypes(symbolname string) []binance.OrderType {
// 	res := make([]binance.OrderType, 0)
// 	symbol := exchangeInfo.Symbols.GetSymbol(symbolname)
// 	for _, orderType := range symbol.OrderTypes {
// 		res = append(res, binance.OrderType(orderType))
// 	}
// 	return res
// }

// func (exchangeInfo *ExchangeInfo) GetPermissions(symbolname string) []string {
// 	return exchangeInfo.Symbols.GetSymbol(symbolname).Permissions
// }
