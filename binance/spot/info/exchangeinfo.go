package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
	symbols_info "github.com/fr0ster/go-trading-utils/binance/spot/info/symbols"
	symbol_info "github.com/fr0ster/go-trading-utils/binance/spot/info/symbols/symbol"
)

type ExchangeInfo struct {
	exchangeInfo *binance.ExchangeInfo
	Symbols      *symbols_info.Symbols
}

func NewExchangeInfo(client *binance.Client) (*ExchangeInfo, error) {
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
