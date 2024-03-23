package bookticker

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
)

func Init(btt *bookticker_types.BookTickerBTree, api_key, secret_key, symbolname string, UseTestnet bool) (err error) {
	futures.UseTestnet = UseTestnet
	client := futures.NewClient(api_key, secret_key)
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	for _, bookTicker := range bookTickerList {
		bookTicker, err := bookticker_types.Binance2BookTicker(bookTicker)
		if err != nil {
			return err
		}
		btt.Set(bookTicker)
	}
	return nil
}
